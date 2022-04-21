package echoadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/labstack/echo/v4"
)

// EchoLambdaV2 makes it easy to send API Gateway proxy V2 events to a echo.Echo.
// The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type EchoLambdaV2 struct {
	core.RequestAccessorV2

	Echo *echo.Echo
}

// NewV2 creates a new instance of the EchoLambda object.
// Receives an initialized *echo.Echo object - normally created with echo.New().
// It returns the initialized instance of the EchoLambdaV2 object.
func NewV2(e *echo.Echo) *EchoLambdaV2 {
	return &EchoLambdaV2{Echo: e}
}

// Proxy receives an API Gateway proxy V2 event, transforms it into an http.Request
// object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (e *EchoLambdaV2) Proxy(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	echoRequest, err := e.ProxyEventToHTTPRequest(req)
	return e.proxyInternal(echoRequest, err)
}

// ProxyWithContext receives context and an API Gateway proxy V2 event,
// transforms them into an http.Request object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (e *EchoLambdaV2) ProxyWithContext(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	echoRequest, err := e.EventToRequestWithContext(ctx, req)
	return e.proxyInternal(echoRequest, err)
}

func (e *EchoLambdaV2) proxyInternal(req *http.Request, err error) (events.APIGatewayV2HTTPResponse, error) {

	if err != nil {
		return core.GatewayTimeoutV2(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := core.NewProxyResponseWriterV2()
	e.Echo.ServeHTTP(http.ResponseWriter(respWriter), req)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeoutV2(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
