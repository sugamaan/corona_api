package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
	"log"
	"os"
)

type Database struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Name     string `json:"name,omitempty"`
	Charset  string `json:"charset,omitempty"`
}

type txAdmin struct {
	*sql.DB
}

type Service struct {
	tx txAdmin
}

func ConnectDb() (*sql.DB, error) {
	// パラメータストア接続
	svc := ssm.New(
		session.Must(session.NewSession()),
		aws.NewConfig().WithRegion(os.Getenv("REGION")),
	)

	res, err := svc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(os.Getenv("DB_CONNECTION_SETTING")),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		panic(err)
	}
	database := Database{}
	err = json.Unmarshal([]byte(*res.Parameter.Value), &database)
	if err != nil {
		panic(err)
	}

	// DB接続
	dbConfig := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", database.User, database.Password, database.Host, database.Port, database.Name, database.Charset)
	if err != nil {
		log.Println(dbConfig)
		panic(err)
	}
	db, err := sql.Open("mysql", dbConfig)
	if err != nil {
		log.Println(dbConfig)
		return nil, err
	}

	// SQLのクエリログを取得
	loggerAdapter := zerologadapter.New(zerolog.New(os.Stdout))
	db = sqldblogger.OpenDriver(dbConfig, db.Driver(), loggerAdapter)

	return db, nil
}

func (t *txAdmin) Transaction(ctx context.Context, f func(ctx context.Context) (err error)) error {
	tx, err := t.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := f(ctx); err != nil {
		return fmt.Errorf("query faild: %w", err)
	}
	return tx.Commit()
}
