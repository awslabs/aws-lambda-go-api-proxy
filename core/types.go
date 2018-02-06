package core

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Returns a dafault Gateway Timeout (504) response
func GatewayTimeout() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusGatewayTimeout}
}

func NewLoggedError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Println(err.Error())
	return err
}
