package core

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func FunctionURLTimeout() events.LambdaFunctionURLResponse {
	return events.LambdaFunctionURLResponse{StatusCode: http.StatusGatewayTimeout}
}
