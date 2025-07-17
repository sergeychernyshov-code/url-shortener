package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	dynamo    *dynamodb.DynamoDB
	tableName = os.Getenv("DYNAMO_TABLE")
	letters   = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func init() {
	rand.Seed(time.Now().UnixNano())
	sess := session.Must(session.NewSession())
	dynamo = dynamodb.New(sess)
}

type ShortenRequest struct {
	URL string `json:"url"`
}

func randomCode(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "POST":
		// Handle /shorten
		var body ShortenRequest
		err := json.Unmarshal([]byte(req.Body), &body)
		if err != nil || body.URL == "" {
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request"}, nil
		}

		code := randomCode(6)

		_, err = dynamo.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"code":     {S: aws.String(code)},
				"long_url": {S: aws.String(body.URL)},
			},
		})
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Database error"}, nil
		}

		response := fmt.Sprintf(`{"short_url":"https://%s/prod/%s"}`, req.Headers["Host"], code)
		return events.APIGatewayProxyResponse{StatusCode: 200, Body: response}, nil

	case "GET":
		// Handle /{code}
		code := req.PathParameters["code"]

		res, err := dynamo.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"code": {S: aws.String(code)},
			},
		})
		if err != nil || res.Item == nil {
			return events.APIGatewayProxyResponse{StatusCode: 404, Body: "Not found"}, nil
		}

		longURL := *res.Item["long_url"].S

		return events.APIGatewayProxyResponse{
			StatusCode: 301,
			Headers: map[string]string{
				"Location": longURL,
			},
		}, nil

	default:
		return events.APIGatewayProxyResponse{StatusCode: 405, Body: "Method not allowed"}, nil
	}
}

func main() {
	lambda.Start(handler)
}
