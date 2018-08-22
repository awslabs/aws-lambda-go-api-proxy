package handlerfunc

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

type HandlerFuncAdapter struct {
	core.RequestAccessor
	handler http.Handler
}

func New(handlerFunc http.HandlerFunc) *HandlerFuncAdapter {
	return &HandlerFuncAdapter{
		handler: handlerFunc,
	}
}

func NewHandler(handler http.Handler) *HandlerFuncAdapter {
	return &HandlerFuncAdapter{
		handler: handler,
	}
}

func (h *HandlerFuncAdapter) Proxy(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(event)
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
