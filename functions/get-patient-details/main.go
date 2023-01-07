package main

import (
	"corona-api/src/middleware"
	"corona-api/src/modules/common"
	"corona-api/src/modules/date"
	"corona-api/src/modules/patient"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"strconv"
)

const (
	StartDateOfCountingPatientDetails = 20200509
)

type PatientDetailParams struct {
	area      string
	startDate uint32
	endDate   uint32
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

	todayInt, err := date.GetToday()
	if err != nil {
		return PatientDetailParams{}, fmt.Errorf("date.GetToday(): todayInt: %v, %v", todayInt, err)
	}

	if startDateInt > endDateInt || startDateInt < StartDateOfCountingPatientDetails || startDateInt >= todayInt || endDateInt < StartDateOfCountingPatientDetails || endDateInt >= todayInt {
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
// @router /patient/details/ [get]
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// クエリパラメーター取得
	patientDetailParams, err := getParams(request)
	if err != nil {
		return common.APIGatewayProxyErrorResponse(err, common.BadRequestMessage, http.StatusBadRequest)
	}

	// DB接続
	db, err := middleware.ConnectDb()
	defer db.Close()
	if err != nil {
		return common.APIGatewayProxyErrorResponse(err, common.InternalServerErrorMessage, http.StatusInternalServerError)
	}

	// SQLでデータを取得
	patientDetails, err := patient.GetPatientDetailsByPeriodAndArea(db, patientDetailParams.area, patientDetailParams.startDate, patientDetailParams.endDate)
	if err != nil {
		return common.APIGatewayProxyErrorResponse(err, common.InternalServerErrorMessage, http.StatusInternalServerError)
	}

	// レスポンス作成
	bytes, err := patient.GeneratePatientDetailsResponse(patientDetails)
	if err != nil {
		return common.APIGatewayProxyErrorResponse(err, common.InternalServerErrorMessage, http.StatusInternalServerError)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bytes),
	}, nil
}

func main() {
	lambda.Start(handler)
}
