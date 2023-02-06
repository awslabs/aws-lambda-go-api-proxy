package echoadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/labstack/echo/v4"
)

// EchoLambdaALB makes it easy to send ALB proxy events to a echo.Echo.
// The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type EchoLambdaALB struct {
	core.RequestAccessorALB

	Echo *echo.Echo
}

// NewAPI creates a new instance of the EchoLambdaAPI object.
// Receives an initialized *echo.Echo object - normally created with echo.New().
// It returns the initialized instance of the EchoLambdaALB object.
func NewALB(e *echo.Echo) *EchoLambdaALB {
	return &EchoLambdaALB{Echo: e}
}

// Proxy receives an ALB event, transforms it into an http.Request
// object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (e *EchoLambdaALB) Proxy(req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	echoRequest, err := e.ProxyEventToHTTPRequest(req)
	return e.proxyInternal(echoRequest, err)
}

// ProxyWithContext receives context and an ALB event,
// transforms them into an http.Request object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (e *EchoLambdaALB) ProxyWithContext(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	echoRequest, err := e.EventToRequestWithContext(ctx, req)
	return e.proxyInternal(echoRequest, err)
}

func (e *EchoLambdaALB) proxyInternal(req *http.Request, err error) (events.ALBTargetGroupResponse, error) {

	if err != nil {
		return core.GatewayTimeoutALB(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := core.NewProxyResponseWriterALB()
	e.Echo.ServeHTTP(http.ResponseWriter(respWriter), req)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeoutALB(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
