package core

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// GatewayTimeout returns Gateway Timeout (504) response
func GatewayTimeout() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusGatewayTimeout}
}

// InternalServerError returns Internal Server Error (500) response
func InternalServerError() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}
}

// NewLoggedError generates a new error and logs it to stdout
func NewLoggedError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Println(err.Error())
	return err
}
