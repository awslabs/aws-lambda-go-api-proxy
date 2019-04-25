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

func (h *HandlerAdapter) Proxy(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return h.ProxyWithContext(context.Background(), req)
}

func (h *HandlerAdapter) ProxyWithContext(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(ctx, event)
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
