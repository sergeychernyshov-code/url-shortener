package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
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

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func randomCode(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func unauthorizedResponse() events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusUnauthorized,
		Body:       "Unauthorized",
		Headers: map[string]string{
			"WWW-Authenticate": `Bearer realm="URL Shortener"`,
		},
	}
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	authHeader := req.Headers["authorization"]
	expectedToken := os.Getenv("API_AUTH_TOKEN")

	if !strings.HasPrefix(authHeader, "Bearer ") || authHeader != "Bearer "+expectedToken {
		return unauthorizedResponse(), nil
	}

	path := req.RawPath
	method := req.RequestContext.HTTP.Method

	switch method {
	case http.MethodPost:
		// Only accept POST on /shorten
		if !strings.HasSuffix(path, "/shorten") {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusNotFound,
				Body:       "Not found",
			}, nil
		}

		var body ShortenRequest
		if err := json.Unmarshal([]byte(req.Body), &body); err != nil || body.URL == "" {
			return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusBadRequest, Body: "Invalid request"}, nil
		}

		code := randomCode(6)

		_, err := dynamo.PutItem(&dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item: map[string]*dynamodb.AttributeValue{
				"code":     {S: aws.String(code)},
				"long_url": {S: aws.String(body.URL)},
			},
		})
		if err != nil {
			return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError, Body: "Database error"}, nil
		}

		shortURL := fmt.Sprintf("https://%s/%s", req.Headers["host"], code)
		response := ShortenResponse{ShortURL: shortURL}
		respBody, _ := json.Marshal(response)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(respBody),
		}, nil

	case http.MethodGet:
		// Expecting GET /{code}
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) != 1 {
			return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusBadRequest, Body: "Invalid URL"}, nil
		}
		code := parts[0]

		res, err := dynamo.GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"code": {S: aws.String(code)},
			},
		})
		if err != nil || res.Item == nil {
			return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNotFound, Body: "Not found"}, nil
		}

		longURL := *res.Item["long_url"].S

		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusMovedPermanently,
			Headers:    map[string]string{"Location": longURL},
		}, nil

	default:
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusMethodNotAllowed, Body: "Method not allowed"}, nil
	}
}

func main() {
	lambda.Start(handler)
}
