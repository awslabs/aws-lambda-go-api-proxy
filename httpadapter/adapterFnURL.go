package httpadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

type HandlerAdapterFnURL struct {
	core.RequestAccessorFnURL
	handler http.Handler
}

func NewFunctionURL(handler http.Handler) *HandlerAdapterFnURL {
	return &HandlerAdapterFnURL{
		handler: handler,
	}
}

// Proxy receives an ALB Target Group proxy event, transforms it into an http.Request
// object, and sends it to the http.HandlerFunc for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (h *HandlerAdapterFnURL) Proxy(event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(event)
	return h.proxyInternal(req, err)
}

// ProxyWithContext receives context and an API Gateway proxy event,
// transforms them into an http.Request object, and sends it to the echo.Echo for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (h *HandlerAdapterFnURL) ProxyWithContext(ctx context.Context, event events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	req, err := h.EventToRequestWithContext(ctx, event)
	return h.proxyInternal(req, err)
}

func (h *HandlerAdapterFnURL) proxyInternal(req *http.Request, err error) (events.LambdaFunctionURLResponse, error) {
	if err != nil {
		return core.FunctionURLTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewFunctionURLResponseWriter()
	h.handler.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		return core.FunctionURLTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return resp, nil
}
