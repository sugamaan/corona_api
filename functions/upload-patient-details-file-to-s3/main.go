package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-resty/resty/v2"
	"log"
	"os"
	"time"
)

const (
	Failure               = 0
	Success               = 1
	Covid19JapanAllURL    = "https://opendata.corona.go.jp/api/Covid19JapanAll"
	S3ObjectKeyTimeFormat = "20060102150405"
)

type Response struct {
	Status    int    `json:"Status"`
	ObjectKey string `json:"ObjectKey"`
}

func handler() (Response, error) {
	// httpリクエストでレスポンスを取得
	c := resty.New()
	res, err := c.SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() != 200
		}).
		R().
		Get(Covid19JapanAllURL)
	file := res.Body()
	reader := bytes.NewReader(file)

	// 取得した情報をS3へ保存
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(os.Getenv("REGION"))},
	})
	if err != nil {
		log.Println(err)
		return Response{Status: Failure}, err
	}

	patientDetailsFileBucketName := fmt.Sprintf("patient-details-file-%s", os.Getenv("ENV"))

	// 現在時刻をオブジェクトキーに設定
	jst, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		log.Println(err)
		return Response{Status: Failure}, err
	}
	now := time.Now().In(jst)
	objectKey := now.Format(S3ObjectKeyTimeFormat)

	upload := s3manager.NewUploader(sess)
	_, err = upload.Upload(&s3manager.UploadInput{
		Bucket: aws.String(patientDetailsFileBucketName),
		Key:    aws.String(objectKey),
		Body:   reader,
	})
	if err != nil {
		log.Println(err)
		return Response{Status: Failure}, err
	}

	return Response{
		Status:    Success,
		ObjectKey: objectKey,
	}, nil
}

func main() {
	lambda.Start(handler)
}
