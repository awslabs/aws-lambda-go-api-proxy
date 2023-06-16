package core

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// GatewayTimeoutFnURL returns a dafault Gateway Timeout (504) response
func GatewayTimeoutFnURL() events.LambdaFunctionURLResponse {
	return events.LambdaFunctionURLResponse{StatusCode: http.StatusGatewayTimeout}
}
