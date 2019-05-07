// Packge echolambda add Echo support for the aws-severless-go-api library.
// Uses the core package behind the scenes and exposes the New method to
// get a new instance and Proxy method to send request to the echo.Echo
package echoadapter

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/labstack/echo"
)

// EchoLambda makes it easy to send API Gateway proxy events to a echo.Echo.
// The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type EchoLambda struct {
	core.RequestAccessor

	Echo *echo.Echo
}

// New creates a new instance of the EchoLambda object.
// Receives an initialized *echo.Echo object - normally created with echo.New().
// It returns the initialized instance of the EchoLambda object.
func New(e *echo.Echo) *EchoLambda {
	return &EchoLambda{Echo: e}
}

// Proxy receives an API Gateway proxy event, transforms it into an http.Request
// object, and sends it to the echo.Echo for routing.
// It returns a proxy response object gneerated from the http.ResponseWriter.
func (e *EchoLambda) Proxy(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	eRequest, err := e.ProxyEventToHTTPRequest(req)

	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := core.NewProxyResponseWriter()
	e.Echo.ServeHTTP(http.ResponseWriter(respWriter), eRequest)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
