package main

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type stubDynamo struct {
	putInput  *dynamodb.PutItemInput
	getInput  *dynamodb.GetItemInput
	putErr    error
	getOutput *dynamodb.GetItemOutput
	getErr    error
}

func (s *stubDynamo) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	s.putInput = input
	return &dynamodb.PutItemOutput{}, s.putErr
}

func (s *stubDynamo) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	s.getInput = input
	if s.getOutput != nil {
		return s.getOutput, s.getErr
	}
	return &dynamodb.GetItemOutput{}, s.getErr
}

func useStubDynamo(t *testing.T, stub *stubDynamo) {
	t.Helper()
	original := dynamoClient
	dynamoClient = stub
	t.Cleanup(func() {
		dynamoClient = original
	})
}

func setTableName(t *testing.T, name string) {
	t.Helper()
	original := tableName
	tableName = name
	t.Cleanup(func() {
		tableName = original
	})
}

func TestRandomCodeGeneratesExpectedOutput(t *testing.T) {
	rand.Seed(1)
	code := randomCode(8)
	if len(code) != 8 {
		t.Fatalf("expected code length 8, got %d", len(code))
	}

	validCharacters := string(letters)
	for _, r := range code {
		if !strings.ContainsRune(validCharacters, r) {
			t.Fatalf("code contains invalid rune %q", r)
		}
	}

	next := randomCode(8)
	if code == next {
		t.Fatalf("expected subsequent random codes to differ")
	}
}

func TestUnauthorizedResponse(t *testing.T) {
	resp := unauthorizedResponse()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if resp.Body != "Unauthorized" {
		t.Fatalf("unexpected body: %s", resp.Body)
	}
	if resp.Headers["WWW-Authenticate"] == "" {
		t.Fatalf("expected WWW-Authenticate header to be set")
	}
}

func TestHandlerUnauthorizedWithoutHeader(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	resp, err := handler(context.Background(), events.APIGatewayV2HTTPRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestHandlerUnauthorizedWrongToken(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer nope",
		},
	}
	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestHandlerPostInvalidPath(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")
	useStubDynamo(t, &stubDynamo{})

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
		},
		RawPath: "/other",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost},
		},
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestHandlerPostInvalidBody(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")
	useStubDynamo(t, &stubDynamo{})

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
		},
		RawPath: "/shorten",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost},
		},
		Body: "{}",
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestHandlerPostDatabaseError(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")
	stub := &stubDynamo{putErr: errors.New("boom")}
	useStubDynamo(t, stub)

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
			"host":          "example.com",
		},
		RawPath: "/shorten",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost},
		},
		Body: `{"url":"https://example.com"}`,
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestHandlerPostSuccess(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")
	rand.Seed(2)
	stub := &stubDynamo{}
	useStubDynamo(t, stub)

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
			"host":          "example.com",
		},
		RawPath: "/shorten",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost},
		},
		Body: `{"url":"https://example.com"}`,
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if ct := resp.Headers["Content-Type"]; ct != "application/json" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}

	var body ShortenResponse
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body.ShortURL == "" {
		t.Fatalf("short url should not be empty")
	}

	if stub.putInput == nil {
		t.Fatalf("expected PutItem to be invoked")
	}

	code := aws.StringValue(stub.putInput.Item["code"].S)
	if !strings.HasSuffix(body.ShortURL, "/"+code) {
		t.Fatalf("short url %q does not include generated code %q", body.ShortURL, code)
	}
	if aws.StringValue(stub.putInput.Item["long_url"].S) != "https://example.com" {
		t.Fatalf("unexpected stored long url: %s", aws.StringValue(stub.putInput.Item["long_url"].S))
	}
}

func TestHandlerGetInvalidPath(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")
	useStubDynamo(t, &stubDynamo{})

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
		},
		RawPath: "/a/b",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet},
		},
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestHandlerGetSuccess(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")

	item := map[string]*dynamodb.AttributeValue{
		"code":     {S: aws.String("abc123")},
		"long_url": {S: aws.String("https://example.com")},
	}
	stub := &stubDynamo{getOutput: &dynamodb.GetItemOutput{Item: item}}
	useStubDynamo(t, stub)

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
		},
		RawPath: "/abc123",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet},
		},
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected status %d, got %d", http.StatusMovedPermanently, resp.StatusCode)
	}
	if loc := resp.Headers["Location"]; loc != "https://example.com" {
		t.Fatalf("expected redirect to stored url, got %q", loc)
	}
	if stub.getInput == nil || aws.StringValue(stub.getInput.Key["code"].S) != "abc123" {
		t.Fatalf("expected GetItem to be called with code abc123")
	}
}

func TestHandlerGetNotFound(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")
	stub := &stubDynamo{getOutput: &dynamodb.GetItemOutput{}}
	useStubDynamo(t, stub)

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
		},
		RawPath: "/missing",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet},
		},
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestHandlerGetDatabaseError(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")
	stub := &stubDynamo{getErr: errors.New("boom")}
	useStubDynamo(t, stub)

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
		},
		RawPath: "/whatever",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet},
		},
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	t.Setenv("API_AUTH_TOKEN", "secret")
	setTableName(t, "urls")
	useStubDynamo(t, &stubDynamo{})

	req := events.APIGatewayV2HTTPRequest{
		Headers: map[string]string{
			"authorization": "Bearer secret",
		},
		RawPath: "/shorten",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPut},
		},
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}
