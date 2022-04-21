package ginadapter

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gin-gonic/gin"
)

// GinLambdaV2 makes it easy to send API Gateway proxy V2 events to a Gin
// Engine. The library transforms the proxy event into an HTTP request and then
// creates a proxy response object from the http.ResponseWriter
type GinLambdaV2 struct {
	core.RequestAccessorV2

	ginEngine *gin.Engine
}

// NewV2 creates a new instance of the GinLambdaV2 object.
// Receives an initialized *gin.Engine object - normally created with gin.Default().
// It returns the initialized instance of the GinLambdaV2 object.
func NewV2(gin *gin.Engine) *GinLambdaV2 {
	return &GinLambdaV2{ginEngine: gin}
}

// Proxy receives an API Gateway proxy V2 event, transforms it into an http.Request
// object, and sends it to the gin.Engine for routing.
// It returns an http response object generated from the http.ResponseWriter.
func (g *GinLambdaV2) Proxy(req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	ginRequest, err := g.ProxyEventToHTTPRequest(req)
	return g.proxyInternal(ginRequest, err)
}

// ProxyWithContext receives context and an API Gateway proxy V2 event,
// transforms them into an http.Request object, and sends it to the gin.Engine for routing.
// It returns an http response object generated from the http.ResponseWriter.
func (g *GinLambdaV2) ProxyWithContext(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	ginRequest, err := g.EventToRequestWithContext(ctx, req)
	return g.proxyInternal(ginRequest, err)
}

func (g *GinLambdaV2) proxyInternal(req *http.Request, err error) (events.APIGatewayV2HTTPResponse, error) {

	if err != nil {
		return core.GatewayTimeoutV2(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	respWriter := core.NewProxyResponseWriterV2()
	g.ginEngine.ServeHTTP(http.ResponseWriter(respWriter), req)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeoutV2(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}
