package core

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func GatewayTimeoutALB() events.ALBTargetGroupResponse {
	return events.ALBTargetGroupResponse{StatusCode: http.StatusGatewayTimeout}
}
