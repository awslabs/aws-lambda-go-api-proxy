package httpadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

type HandlerAdapter struct {
	core.RequestAccessor
	handler http.Handler
}

func New(handler http.Handler) *HandlerAdapter {
	return &HandlerAdapter{
		handler: handler,
	}
}

// Proxy receives an API Gateway proxy event, transforms it into an http.Request
// object, and sends it to the http.HandlerFunc for routing.
// It returns a proxy response object generated from the http.Handler.
func (h *HandlerAdapter) Proxy(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(event)
	return h.proxyInternal(req, err)
}

// ProxyWithContext receives context and an API Gateway proxy event,
// transforms them into an http.Request object, and sends it to the http.Handler for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (h *HandlerAdapter) ProxyWithContext(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req, err := h.EventToRequestWithContext(ctx, event)
	return h.proxyInternal(req, err)
}

func (h *HandlerAdapter) proxyInternal(req *http.Request, err error) (events.APIGatewayProxyResponse, error) {
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewProxyResponseWriter()
	h.handler.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return resp, nil
}
