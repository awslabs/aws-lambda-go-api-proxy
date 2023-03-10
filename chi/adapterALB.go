package chiadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/go-chi/chi/v5"
)

// ChiLambdaALB makes it easy to send ALB proxy events to a chi.Mux.
// The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type ChiLambdaALB struct {
	core.RequestAccessorALB

	Chi *chi.Mux
}

// NewALB creates a new instance of the ChiLambdaALB object.
// Receives an initialized *chi.Mux object - normally created with chi.New().
// It returns the initialized instance of the ChiLambdaALB object.
func NewALB(chi *chi.Mux) *ChiLambdaALB {
	return &ChiLambdaALB{Chi: chi}
}

// Proxy receives an ALB proxy event, transforms it into an http.Request
// object, and sends it to the chi.Mux for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (c *ChiLambdaALB) Proxy(req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	chiRequest, err := c.ProxyEventToHTTPRequest(req)
	return c.proxyInternal(chiRequest, err)
}

// ProxyWithContext receives an ALB proxy event, transforms it into an http.Request
// object, and sends it to the chi.Mux for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (c *ChiLambdaALB) ProxyWithContext(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	chiRequest, err := c.EventToRequestWithContext(ctx, req)
	return c.proxyInternal(chiRequest, err)
}

func (c *ChiLambdaALB) proxyInternal(req *http.Request, err error) (events.ALBTargetGroupResponse, error) {

	if err != nil {
		return core.GatewayTimeoutALB(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := core.NewProxyResponseWriterALB()
	c.Chi.ServeHTTP(http.ResponseWriter(respWriter), req)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeoutALB(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
