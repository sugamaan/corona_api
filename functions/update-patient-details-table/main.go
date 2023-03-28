package main

import (
	"corona-api/src/middleware"
	"corona-api/src/modules/patient"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
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

func handler(event Event) (Response, error) {
	// DB接続
	db, err := middleware.ConnectDb()
	defer db.Close()
	if err != nil {
		return Response{Status: Failure}, err
	}

	// セッション
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(os.Getenv("REGION"))},
	})
	if err != nil {
		return Response{Status: Failure}, err
	}

	// S3からJSONファイルを取得
	svc := s3.New(sess)
	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(fmt.Sprintf("patient-details-file-%s", os.Getenv("ENV"))),
		Key:    aws.String(event.ObjectKey),
	})
	if err != nil {
		return Response{Status: Failure}, err
	}
	file2, err := io.ReadAll(obj.Body)
	if err != nil {
		return Response{Status: Failure}, err
	}

	var patientDetailsResponse PatientDetailsResponse
	err = json.Unmarshal(file2, &patientDetailsResponse)
	if err != nil {
		return Response{Status: Failure}, err
	}

	// インサート用の構造体へ変換する
	var PatientDetails []patient.Detail
	for _, item := range patientDetailsResponse.ItemList {
		result := strings.Replace(item.Date, "-", "", -1)
		date, err := strconv.Atoi(result)
		if err != nil {
			return Response{Status: Failure}, err
		}
		npatients, err := strconv.Atoi(item.Npatients)
		if err != nil {
			return Response{Status: Failure}, err
		}
		pd := patient.Detail{
			Date:    uint32(date),
			Area:    item.NameJp,
			Value:   uint32(npatients),
			Country: "日本",
		}
		PatientDetails = append(PatientDetails, pd)
	}

	// DBへ保存
	err = patient.InsertPatientDetails(db, PatientDetails)
	if err != nil {
		return Response{Status: Failure}, err
	}

	return Response{
		Status: Success,
	}, nil
}

func main() {
	lambda.Start(handler)
}
