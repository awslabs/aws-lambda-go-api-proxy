package negroniadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/urfave/negroni"
)

type NegroniAdapter struct {
	core.RequestAccessor
	n *negroni.Negroni
}

func New(n *negroni.Negroni) *NegroniAdapter {
	return &NegroniAdapter{
		n: n,
	}
}

func (h *NegroniAdapter) Proxy(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return h.ProxyWithContext(context.Background(), req)
}

func (h *NegroniAdapter) ProxyWithContext(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(ctx, event)
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewProxyResponseWriter()
	h.n.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return resp, nil
}
