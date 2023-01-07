package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
	"net/http"
	"net/url"
	"os"
)

type payload struct {
	Text string `json:"text"`
}

func handler() {
	// パラメータストア接続
	svc := ssm.New(
		session.Must(session.NewSession()),
		aws.NewConfig().WithRegion(os.Getenv("REGION")),
	)

	res, err := svc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String("NotifyExecutionOfPatientDetailsWebhookUrl"),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		panic(err)
	}

	// slackへ通知
	hookUrl := *res.Parameter.Value
	message, err := json.Marshal(payload{
		Text: "定期実行に失敗しました",
	})
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.PostForm(hookUrl, url.Values{"payload": {string(message)}})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func main() {
	lambda.Start(handler)
}
