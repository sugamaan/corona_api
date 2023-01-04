package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"hello-world/src/middleware"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	BadRequestMessage  = "リクエストパラメータが不正です"
	AvailableStartDate = 20200509
)

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
	HttpStatus   int    `json:"httpStatus"`
}

type PatientDetailParams struct {
	area      string
	startDate uint32
	endDate   uint32
}

type PatientDetail struct {
	Date    uint32
	Area    string
	Value   uint32
	Country string
}

func getParams(request events.APIGatewayProxyRequest) (PatientDetailParams, error) {
	area := request.QueryStringParameters["area"]
	startDate := request.QueryStringParameters["start_date"]
	endDate := request.QueryStringParameters["end_date"]

	if area == "" || startDate == "" || endDate == "" {
		return PatientDetailParams{}, fmt.Errorf("missing required parameter: area: %v, startDateInt: %v, endDateInt: %v", area, startDate, endDate)
	}

	startDateInt, err := strconv.Atoi(startDate)
	if err != nil {
		return PatientDetailParams{}, fmt.Errorf("strconv.Atoi(startDate): startDate: %v, %v", startDate, err)
	}

	endDateInt, err := strconv.Atoi(endDate)
	if err != nil {
		return PatientDetailParams{}, fmt.Errorf("strconv.Atoi(endDate): endDate: %v, %v", endDate, err)
	}

	jst, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		return PatientDetailParams{}, fmt.Errorf("time.LoadLocation error: %v", err)
	}
	now := time.Now().In(jst)
	today := now.Format("20060102")
	todayInt, err := strconv.Atoi(today)
	if err != nil {
		return PatientDetailParams{}, fmt.Errorf("strconv.Atoi(today): today: %v, %v", todayInt, err)
	}

	if startDateInt > endDateInt || startDateInt < AvailableStartDate || startDateInt >= todayInt || endDateInt < AvailableStartDate || endDateInt >= todayInt {
		return PatientDetailParams{}, fmt.Errorf("invalid specified period: startDateInt: %v, endDateInt: %v", startDate, endDate)
	}

	return PatientDetailParams{
		area,
		uint32(startDateInt),
		uint32(endDateInt),
	}, nil
}

// @summary	感染者数詳細リスト取得
// @description 2020/05/09から前日までの指定都道府県の感染者数情報を取得する
// @tags Patients
// @accept json
// @produce json
// @param start_date query int ture "開始日" example(20230101)
// @param end_date query int ture "終了日" example(20230102)
// @param area query string ture "都道府県名" example("北海道")
// @Success 200
// @failure 400
// @failure 500
// @router /patients/details/ [get]
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// クエリパラメーター取得
	patientDetailParams, err := getParams(request)
	if err != nil {
		// ロギング
		log.Println(err)

		// カスタムレスポンス
		res := &ErrorResponse{
			ErrorMessage: BadRequestMessage,
			HttpStatus:   http.StatusBadRequest,
		}
		body, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
		}
		return events.APIGatewayProxyResponse{
			Body:       string(body),
			StatusCode: http.StatusBadRequest,
		}, err
	}

	// DB接続
	db, err := middleware.ConnectDb()
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, err
	}

	// SQL作成
	rows, err := db.Query("SELECT  date, area, value, country FROM patient_details WHERE area = ? AND date BETWEEN ? AND ?", patientDetailParams.area, patientDetailParams.startDate, patientDetailParams.endDate)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, err
	}
	defer rows.Close()

	var patientDetails []PatientDetail
	for rows.Next() {
		var patientDetail PatientDetail
		err := rows.Scan(&patientDetail.Date, &patientDetail.Area, &patientDetail.Value, &patientDetail.Country)
		if err != nil {
			log.Fatal(err)
		}
		patientDetails = append(patientDetails, patientDetail)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	// レスポンスの形に整形
	body := map[string]interface{}{}

	// エリア
	body["area"] = patientDetails[0].Area

	// 日付データを作成
	for _, pd := range patientDetails {
		dateString := strconv.Itoa(int(pd.Date))
		body[dateString] = pd.Value
	}

	// 合計データを作成
	var sum uint32
	for _, pd := range patientDetails {
		sum += pd.Value
	}
	body["sum"] = sum

	// 平均データを作成
	average := int(sum) / len(patientDetails)
	body["average"] = average

	// JSONにして返却
	bytes, err := json.Marshal(body)
	if err != nil {
		fmt.Println("JSON marshal error: ", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bytes),
	}, nil
}

func main() {
	lambda.Start(handler)
}
