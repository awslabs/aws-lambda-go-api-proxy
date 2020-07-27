// Package irisLambda add Iris support for the aws-serverless-go-api library.
// Uses the core package behind the scenes and exposes the New method to
// get a new instance and Proxy method to send request to the iris.Application.
package irisadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/kataras/iris/v12"
)

// IrisLambda makes it easy to send API Gateway proxy events to a iris.Application.
// The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type IrisLambda struct {
	core.RequestAccessor

	application *iris.Application
}

// New creates a new instance of the IrisLambda object.
// Receives an initialized *iris.Application object - normally created with iris.Default().
// It returns the initialized instance of the IrisLambda object.
func New(app *iris.Application) *IrisLambda {
	return &IrisLambda{application: app}
}

// Proxy receives an API Gateway proxy event, transforms it into an http.Request
// object, and sends it to the iris.Application for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (i *IrisLambda) Proxy(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	irisRequest, err := i.ProxyEventToHTTPRequest(req)
	return i.proxyInternal(irisRequest, err)
}

// ProxyWithContext receives context and an API Gateway proxy event,
// transforms them into an http.Request object, and sends it to the iris.Application for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (i *IrisLambda) ProxyWithContext(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	irisRequest, err := i.EventToRequestWithContext(ctx, req)
	return i.proxyInternal(irisRequest, err)
}

func (i *IrisLambda) proxyInternal(req *http.Request, err error) (events.APIGatewayProxyResponse, error) {
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	if err := i.application.Build(); err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Iris set up failed: %v", err)
	}

	respWriter := core.NewProxyResponseWriter()
	i.application.ServeHTTP(http.ResponseWriter(respWriter), req)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
