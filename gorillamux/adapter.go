package gorillamux

import (
	"context"
	"errors"
	"net/http"

	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gorilla/mux"
)

type GorillaMuxAdapter struct {
	RequestAccessor core.RequestAccessor
	RequestAccessorV2 core.RequestAccessorV2
	router *mux.Router
}

func New(router *mux.Router) *GorillaMuxAdapter {
	return &GorillaMuxAdapter{
		router: router,
	}
}

// Proxy receives an API Gateway proxy event or API Gateway V2 event, transforms it into an http.Request
// object, and sends it to the mux.Router for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (h *GorillaMuxAdapter) Proxy(event core.SwitchableAPIGatewayRequest) (*core.SwitchableAPIGatewayResponse, error) {
	if event.Version1() != nil {
		req, err := h.RequestAccessor.ProxyEventToHTTPRequest(*event.Version1())
		return h.proxyInternal(req, err)
	} else if event.Version2() != nil {
		req, err := h.RequestAccessorV2.ProxyEventToHTTPRequest(*event.Version2())
		return h.proxyInternalV2(req, err)
	} else {
		return &core.SwitchableAPIGatewayResponse{}, core.NewLoggedError("Could not convert proxy event to request: %v", errors.New("Unable to determine version "))
	}
}

// ProxyWithContext receives context and an API Gateway proxy event or API Gateway V2 event,
// transforms them into an http.Request object, and sends it to the mux.Router for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (h *GorillaMuxAdapter) ProxyWithContext(ctx context.Context, event core.SwitchableAPIGatewayRequest) (*core.SwitchableAPIGatewayResponse, error) {
	if event.Version1() != nil {
		req, err := h.RequestAccessor.EventToRequestWithContext(ctx, *event.Version1())
		return h.proxyInternal(req, err)
	} else if event.Version2() != nil {
		req, err := h.RequestAccessorV2.EventToRequestWithContext(ctx, *event.Version2())
		return h.proxyInternalV2(req, err)
	} else {
		return &core.SwitchableAPIGatewayResponse{}, core.NewLoggedError("Could not convert proxy event to request: %v", errors.New("Unable to determine version "))
	}
}

func (h *GorillaMuxAdapter) proxyInternal(req *http.Request, err error) (*core.SwitchableAPIGatewayResponse, error) {
	if err != nil {
		timeout := core.GatewayTimeout()
		return core.NewSwitchableAPIGatewayResponseV1(&timeout), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewProxyResponseWriter()
	h.router.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		timeout := core.GatewayTimeout()
		return core.NewSwitchableAPIGatewayResponseV1(&timeout), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return core.NewSwitchableAPIGatewayResponseV1(&resp), nil
}

func (h *GorillaMuxAdapter) proxyInternalV2(req *http.Request, err error) (*core.SwitchableAPIGatewayResponse, error) {
	if err != nil {
		timeout := core.GatewayTimeoutV2()
		return core.NewSwitchableAPIGatewayResponseV2(&timeout), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewProxyResponseWriterV2()
	h.router.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		timeout := core.GatewayTimeoutV2()
		return core.NewSwitchableAPIGatewayResponseV2(&timeout), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return core.NewSwitchableAPIGatewayResponseV2(&resp), nil
}
