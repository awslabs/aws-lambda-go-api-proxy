package gorillamux

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gorilla/mux"
)

type GorillaMuxAdapter struct {
	core.RequestAccessor
	router *mux.Router
}

func New(router *mux.Router) *GorillaMuxAdapter {
	return &GorillaMuxAdapter{
		router: router,
	}
}

func (h *GorillaMuxAdapter) Proxy(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return h.ProxyWithContext(context.Background(), req)
}

func (h *GorillaMuxAdapter) ProxyWithContext(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(ctx, event)
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewProxyResponseWriter()
	h.router.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return resp, nil
}
