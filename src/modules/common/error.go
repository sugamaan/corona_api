package common

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"log"
	"net/http"
)

const (
	BadRequestMessage          = "リクエストパラメータが不正です"
	InternalServerErrorMessage = "サーバーエラーが発生しました。運営にお問合せください"
)

type CustomLambdaErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
	HttpStatus   int    `json:"httpStatus"`
}

func APIGatewayProxyErrorResponse(err error, errorMessage string, httpStatus int) (events.APIGatewayProxyResponse, error) {
	// ロギング
	log.Println(err)

	// レスポンス生成
	res := &CustomLambdaErrorResponse{
		ErrorMessage: errorMessage,
		HttpStatus:   httpStatus,
	}
	body, err := json.Marshal(res)
	if err != nil {
		log.Printf("{\"level\":\"warn\",\"error_message\":\"%s\"}", err.Error())
	}

	return events.APIGatewayProxyResponse{
		Body:       string(body),
		StatusCode: http.StatusBadRequest,
	}, err
}
