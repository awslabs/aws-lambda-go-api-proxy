package core

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func GatewayTimeoutV2() events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusGatewayTimeout}
}
