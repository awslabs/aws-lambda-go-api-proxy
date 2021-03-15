package handlerfunc

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

type HandlerFuncAdapterV2 struct {
	core.RequestAccessorV2
	handlerFunc http.HandlerFunc
}

func NewV2(handlerFunc http.HandlerFunc) *HandlerFuncAdapterV2 {
	return &HandlerFuncAdapterV2{
		handlerFunc: handlerFunc,
	}
}

// Proxy receives an API Gateway proxy event, transforms it into an http.Request
// object, and sends it to the http.HandlerFunc for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (h *HandlerFuncAdapterV2) Proxy(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(event)
	return h.proxyInternal(req, err)
}

// ProxyWithContext receives context and an API Gateway proxy event,
// transforms them into an http.Request object, and sends it to the http.HandlerFunc for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (h *HandlerFuncAdapterV2) ProxyWithContext(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	req, err := h.EventToRequestWithContext(ctx, event)
	return h.proxyInternal(req, err)
}

func (h *HandlerFuncAdapterV2) proxyInternal(req *http.Request, err error) (events.APIGatewayV2HTTPResponse, error) {
	if err != nil {
		return core.GatewayTimeoutV2(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewProxyResponseWriterV2()
	h.handlerFunc.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeoutV2(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return resp, nil
}
