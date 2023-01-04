package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"hello-world/src/middleware"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	Failure = 0
	Success = 1
)

type Event struct {
	ObjectKey string `json:"ObjectKey"`
}

type Response struct {
	Status int `json:"Status"`
}

type PatientDetailsResponse struct {
	ErrorInfo ErrorInfo `json:"errorInfo"`
	ItemList  ItemList  `json:"itemList"`
}

type ErrorInfo struct {
	ErrorFlag    string      `json:"errorFlag"`
	ErrorCode    interface{} `json:"errorCode"`
	ErrorMessage interface{} `json:"errorMessage"`
}

type ItemList []struct {
	Date      string `json:"date"`
	NameJp    string `json:"name_jp"`
	Npatients string `json:"npatients"`
}

type PatientDetail struct {
	Date    uint32
	Area    string
	Value   uint32
	Country string
}

func handler(event Event) (Response, error) {
	// DB接続
	db, err := middleware.ConnectDb()
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("fail connect db", err)
		return Response{Status: Failure}, err
	}

	// セッション
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(os.Getenv("REGION"))},
	})
	if err != nil {
		log.Println(err)
		return Response{Status: Failure}, err
	}

	// S3からJSONファイルを取得
	svc := s3.New(sess)
	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(fmt.Sprintf("patient-details-file-%s", os.Getenv("ENV"))),
		Key:    aws.String(event.ObjectKey),
	})
	if err != nil {
		log.Fatal(err)
	}
	file2, err := io.ReadAll(obj.Body)
	if err != nil {
		log.Fatal(err)
	}

	var patientDetailsResponse PatientDetailsResponse
	err = json.Unmarshal(file2, &patientDetailsResponse)
	if err != nil {
		log.Fatal(err)
	}

	// インサート用の構造体へ変換する
	var PatientDetails []PatientDetail
	for _, item := range patientDetailsResponse.ItemList {
		result := strings.Replace(item.Date, "-", "", -1)
		date, err := strconv.Atoi(result)
		if err != nil {
			log.Fatal(err)
		}
		npatients, err := strconv.Atoi(item.Npatients)
		if err != nil {
			log.Fatal(err)
		}
		pd := PatientDetail{
			Date:    uint32(date),
			Area:    item.NameJp,
			Value:   uint32(npatients),
			Country: "日本",
		}
		PatientDetails = append(PatientDetails, pd)
	}

	fmt.Printf("%+v\n", PatientDetails)

	// DBトランザクション開始
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// DBをTRUNCATE
	patientDetailsTableName := "patient_details"
	_, err = tx.Exec("TRUNCATE " + patientDetailsTableName)
	if err != nil {
		tx.Rollback()
		log.Fatal("truncate table error:", err)
	}

	// DBに保存
	stmt, err := tx.Prepare("INSERT INTO  patient_details (date, area, value, country) VALUES (?,?,?,?)")
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, pd := range PatientDetails {
		_, err := stmt.Exec(pd.Date, pd.Area, pd.Value, pd.Country)
		if err != nil {
			log.Fatal(err)
		}
	}

	// SQLが正常に実行できたらコミット
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	return Response{
		Status: Success,
	}, nil
}

func main() {
	lambda.Start(handler)
}
