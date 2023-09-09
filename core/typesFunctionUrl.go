package core

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func FunctionUrlTimeout() events.LambdaFunctionURLResponse {
	return events.LambdaFunctionURLResponse{StatusCode: http.StatusGatewayTimeout}
}
