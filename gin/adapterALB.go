// Package ginadapter adds Gin support for the aws-severless-go-api library.
// Uses the core package behind the scenes and exposes the New and NewV2 and ALB methods to
// get a new instance and Proxy method to send request to the Gin engine.
package ginadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gin-gonic/gin"
)

// GinLambdaALB makes it easy to send ALB proxy events to a Gin
// Engine. The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type GinLambdaALB struct {
	core.RequestAccessorALB

	ginEngine *gin.Engine
}

// New creates a new instance of the GinLambdaALB object.
// Receives an initialized *gin.Engine object - normally created with gin.Default().
// It returns the initialized instance of the GinLambdaALB object.
func NewALB(gin *gin.Engine) *GinLambdaALB {
	return &GinLambdaALB{ginEngine: gin}
}

// Proxy receives an ALB proxy event, transforms it into an http.Request
// object, and sends it to the gin.Engine for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (g *GinLambdaALB) Proxy(req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	ginRequest, err := g.ProxyEventToHTTPRequest(req)
	return g.proxyInternal(ginRequest, err)
}

// ProxyWithContext receives context and an ALB proxy event,
// transforms them into an http.Request object, and sends it to the gin.Engine for routing.
// It returns a proxy response object generated from the http.ResponseWriter.
func (g *GinLambdaALB) ProxyWithContext(ctx context.Context, req events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	ginRequest, err := g.EventToRequestWithContext(ctx, req)
	return g.proxyInternal(ginRequest, err)
}

func (g *GinLambdaALB) proxyInternal(req *http.Request, err error) (events.ALBTargetGroupResponse, error) {

	if err != nil {
		return core.GatewayTimeoutALB(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := core.NewProxyResponseWriterALB()
	g.ginEngine.ServeHTTP(http.ResponseWriter(respWriter), req)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeoutALB(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
