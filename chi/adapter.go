// Packge chilambda add Chi support for the aws-severless-go-api library.
// Uses the core package behind the scenes and exposes the New method to
// get a new instance and Proxy method to send request to the Chi mux.
package chiadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/go-chi/chi"
)

// ChiLambda makes it easy to send API Gateway proxy events to a Chi
// Mux. The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type ChiLambda struct {
	core.RequestAccessor

	chiMux *chi.Mux
}

// New creates a new instance of the ChiLambda object.
// Receives an initialized *chi.Mux object - normally created with chi.NewRouter().
// It returns the initialized instance of the ChiLambda object.
func New(chi *chi.Mux) *ChiLambda {
	return &ChiLambda{chiMux: chi}
}

// Proxy receives an API Gateway proxy event, transforms it into an http.Request
// object, and sends it to the chi.Mux for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (g *ChiLambda) Proxy(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	chiRequest, err := g.ProxyEventToHTTPRequest(req)
	return g.proxyInternal(chiRequest, err)
}

// ProxyWithContext receives context and an API Gateway proxy event,
// transforms them into an http.Request object, and sends it to the chi.Mux for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (g *ChiLambda) ProxyWithContext(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	chiRequest, err := g.EventToRequestWithContext(ctx, req)
	return g.proxyInternal(chiRequest, err)
}

func (g *ChiLambda) proxyInternal(chiRequest *http.Request, err error) (events.APIGatewayProxyResponse, error) {

	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := core.NewProxyResponseWriter()
	g.chiMux.ServeHTTP(http.ResponseWriter(respWriter), chiRequest)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
