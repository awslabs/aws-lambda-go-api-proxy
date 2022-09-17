// Package core provides utility methods that help convert proxy events
// into an http.Request and http.ResponseWriter
package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

const (
	// APIGwContextHeader is the custom header key used to store the
	// API Gateway context. To access the Context properties use the
	// GetAPIGatewayContext method of the RequestAccessor object.
	ALBTgContextHeader = "X-GoLambdaProxy-AlbTargetGroup-Context"

	// APIGwStageVarsHeader is the custom header key used to store the
	// API Gateway stage variables. To access the stage variable values
	// use the GetAPIGatewayStageVars method of the RequestAccessor object.
	ALBTgStageVarsHeader = "X-GoLambdaProxy-AlbTargetGroup-StageVars"
)

// RequestAccessorALB objects give access to custom API Gateway properties
// in the request.
type RequestAccessorALB struct {
	stripBasePath string
}

// GetAPIGatewayContextV2 extracts the API Gateway context object from a
// request's custom header.
// Returns a populated events.APIGatewayProxyRequestContext object from
// the request.
func (r *RequestAccessorALB) GetALBTargetGroupRequestContext(req *http.Request) (events.ALBTargetGroupRequestContext, error) {
	if req.Header.Get(ALBTgContextHeader) == "" {
		return events.ALBTargetGroupRequestContext{}, errors.New("No context header in request")
	}
	context := events.ALBTargetGroupRequestContext{}
	err := json.Unmarshal([]byte(req.Header.Get(ALBTgContextHeader)), &context)
	if err != nil {
		log.Println("Erorr while unmarshalling context")
		log.Println(err)
		return events.ALBTargetGroupRequestContext{}, err
	}
	return context, nil
}

// StripBasePath instructs the RequestAccessor object that the given base
// path should be removed from the request path before sending it to the
// framework for routing. This is used when API Gateway is configured with
// base path mappings in custom domain names.
func (r *RequestAccessorALB) StripBasePath(basePath string) string {
	if strings.Trim(basePath, " ") == "" {
		r.stripBasePath = ""
		return ""
	}

	newBasePath := basePath
	if !strings.HasPrefix(newBasePath, "/") {
		newBasePath = "/" + newBasePath
	}

	if strings.HasSuffix(newBasePath, "/") {
		newBasePath = newBasePath[:len(newBasePath)-1]
	}

	r.stripBasePath = newBasePath

	return newBasePath
}

// ProxyEventToHTTPRequest converts an API Gateway proxy event into a http.Request object.
// Returns the populated http request with additional two custom headers for the stage variables and API Gateway context.
// To access these properties use the GetAPIGatewayStageVars and GetAPIGatewayContext method of the RequestAccessor object.
func (r *RequestAccessorALB) ProxyEventToHTTPRequest(req events.ALBTargetGroupRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeaderALB(httpRequest, req)
}

// EventToRequestWithContext converts an API Gateway proxy event and context into an http.Request object.
// Returns the populated http request with lambda context, stage variables and APIGatewayProxyRequestContext as part of its context.
// Access those using GetAPIGatewayContextFromContext, GetStageVarsFromContext and GetRuntimeContextFromContext functions in this package.
func (r *RequestAccessorALB) EventToRequestWithContext(ctx context.Context, req events.ALBTargetGroupRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContextALB(ctx, httpRequest, req), nil
}

// EventToRequest converts an API Gateway proxy event into an http.Request object.
// Returns the populated request maintaining headers
func (r *RequestAccessorALB) EventToRequest(req events.ALBTargetGroupRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	path := req.Path

	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if len(req.QueryStringParameters) > 0 {
		values := url.Values{}
		for key, value := range req.QueryStringParameters {
			values.Add(key, value)
		}
		path += "?" + values.Encode()
	}

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.HTTPMethod),
		path,
		bytes.NewReader(decodedBody),
	)

	if err != nil {
		fmt.Printf("Could not convert request %s:%s to http.Request\n", req.HTTPMethod, req.Path)
		log.Println(err)
		return nil, err
	}

	for headerKey, headerValue := range req.Headers {
		for _, val := range strings.Split(headerValue, ",") {
			httpRequest.Header.Add(headerKey, strings.Trim(val, " "))
		}
	}

	httpRequest.RequestURI = httpRequest.URL.RequestURI()

	//for k, v := range req.Headers {
	//	httpRequest.Header.Add(k, v)
	//}

	return httpRequest, nil
}

func addToHeaderALB(req *http.Request, albTgRequest events.ALBTargetGroupRequest) (*http.Request, error) {
	albTgContext, err := json.Marshal(albTgRequest.RequestContext)
	if err != nil {
		log.Println("Could not Marshal API GW context for custom header")
		return req, err
	}
	req.Header.Add(ALBTgContextHeader, string(albTgContext))
	return req, nil
}

func addToContextALB(ctx context.Context, req *http.Request, albTgRequest events.ALBTargetGroupRequest) *http.Request {
	lc, _ := lambdacontext.FromContext(ctx)
	rc := requestContextALB{lambdaContext: lc, gatewayProxyContext: albTgRequest.RequestContext}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

// GetAPIGatewayV2ContextFromContext retrieve APIGatewayProxyRequestContext from context.Context
func GetALBTargetGroupContextFromContext(ctx context.Context) (events.ALBTargetGroupRequestContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextALB)
	return v.gatewayProxyContext, ok
}

// GetRuntimeContextFromContextV2 retrieve Lambda Runtime Context from context.Context
func GetRuntimeContextFromContextALB(ctx context.Context) (*lambdacontext.LambdaContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextALB)
	return v.lambdaContext, ok
}

type requestContextALB struct {
	lambdaContext       *lambdacontext.LambdaContext
	gatewayProxyContext events.ALBTargetGroupRequestContext
}
