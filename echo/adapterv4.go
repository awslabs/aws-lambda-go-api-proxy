// Package echoadapter adds Echo support for the aws-severless-go-api library.
// Uses the core package behind the scenes and exposes the New method to
// get a new instance and Proxy method to send request to the echo.Echo
package echoadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/labstack/echo/v4"
)

// EchoLambdaV4 makes it easy to send API Gateway proxy events to a echo.
// The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type EchoLambdaV4 struct {
	core.RequestAccessor

	Echo *echo.Echo
}

// NewV4 creates a new instance of the EchoLambdaV4 object.
// Receives an initialized *echo.Echo object - normally created with echo.New().
// It returns the initialized instance of the EchoLambda object.
func NewV4(e *echo.Echo) *EchoLambdaV4 {
	return &EchoLambdaV4{Echo: e}
}

// Proxy receives an API Gateway proxy event, transforms it into an http.Request
// object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (e *EchoLambdaV4) Proxy(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	echoRequest, err := e.ProxyEventToHTTPRequest(req)
	return e.proxyInternal(echoRequest, err)
}

// ProxyWithContext receives context and an API Gateway proxy event,
// transforms them into an http.Request object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (e *EchoLambdaV4) ProxyWithContext(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	echoRequest, err := e.EventToRequestWithContext(ctx, req)
	return e.proxyInternal(echoRequest, err)
}

func (e *EchoLambdaV4) proxyInternal(req *http.Request, err error) (events.APIGatewayProxyResponse, error) {

	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := core.NewProxyResponseWriter()
	e.Echo.ServeHTTP(http.ResponseWriter(respWriter), req)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
