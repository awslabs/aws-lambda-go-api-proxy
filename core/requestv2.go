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
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// RequestAccessorV2 objects give access to custom API Gateway properties
// in the request.
type RequestAccessorV2 struct {
	stripBasePath string
}

// GetAPIGatewayContextV2 extracts the API Gateway context object from a
// request's custom header.
// Returns a populated events.APIGatewayProxyRequestContext object from
// the request.
func (r *RequestAccessorV2) GetAPIGatewayContextV2(req *http.Request) (events.APIGatewayV2HTTPRequestContext, error) {
	if req.Header.Get(APIGwContextHeader) == "" {
		return events.APIGatewayV2HTTPRequestContext{}, errors.New("No context header in request")
	}
	context := events.APIGatewayV2HTTPRequestContext{}
	err := json.Unmarshal([]byte(req.Header.Get(APIGwContextHeader)), &context)
	if err != nil {
		log.Println("Erorr while unmarshalling context")
		log.Println(err)
		return events.APIGatewayV2HTTPRequestContext{}, err
	}
	return context, nil
}

// GetAPIGatewayStageVars extracts the API Gateway stage variables from a
// request's custom header.
// Returns a map[string]string of the stage variables and their values from
// the request.
func (r *RequestAccessorV2) GetAPIGatewayStageVars(req *http.Request) (map[string]string, error) {
	stageVars := make(map[string]string)
	if req.Header.Get(APIGwStageVarsHeader) == "" {
		return stageVars, errors.New("No stage vars header in request")
	}
	err := json.Unmarshal([]byte(req.Header.Get(APIGwStageVarsHeader)), &stageVars)
	if err != nil {
		log.Println("Erorr while unmarshalling stage variables")
		log.Println(err)
		return stageVars, err
	}
	return stageVars, nil
}

// StripBasePath instructs the RequestAccessor object that the given base
// path should be removed from the request path before sending it to the
// framework for routing. This is used when API Gateway is configured with
// base path mappings in custom domain names.
func (r *RequestAccessorV2) StripBasePath(basePath string) string {
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
func (r *RequestAccessorV2) ProxyEventToHTTPRequest(req events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeaderV2(httpRequest, req)
}

// EventToRequestWithContext converts an API Gateway proxy event and context into an http.Request object.
// Returns the populated http request with lambda context, stage variables and APIGatewayProxyRequestContext as part of its context.
// Access those using GetAPIGatewayContextFromContext, GetStageVarsFromContext and GetRuntimeContextFromContext functions in this package.
func (r *RequestAccessorV2) EventToRequestWithContext(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContextV2(ctx, httpRequest, req), nil
}

// EventToRequest converts an API Gateway proxy event into an http.Request object.
// Returns the populated request maintaining headers
func (r *RequestAccessorV2) EventToRequest(req events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	path := req.RawPath

	//if RawPath empty is, populate from request context
	if len(path) == 0 {
		path = req.RequestContext.HTTP.Path
	}

	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	serverAddress := "https://" + req.RequestContext.DomainName
	if customAddress, ok := os.LookupEnv(CustomHostVariable); ok {
		serverAddress = customAddress
	}
	path = serverAddress + path

	if len(req.RawQueryString) > 0 {
		path += "?" + req.RawQueryString
	} else if len(req.QueryStringParameters) > 0 {
		values := url.Values{}
		for key, value := range req.QueryStringParameters {
			values.Add(key, value)
		}
		path += "?" + values.Encode()
	}

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.RequestContext.HTTP.Method),
		path,
		bytes.NewReader(decodedBody),
	)

	if err != nil {
		fmt.Printf("Could not convert request %s:%s to http.Request\n", req.RequestContext.HTTP.Method, req.RequestContext.HTTP.Path)
		log.Println(err)
		return nil, err
	}

	for headerKey, headerValue := range req.Headers {
		for _, val := range strings.Split(headerValue, ",") {
			httpRequest.Header.Add(headerKey, strings.Trim(val, " "))
		}
	}

	httpRequest.RequestURI = httpRequest.URL.RequestURI()

	return httpRequest, nil
}

func addToHeaderV2(req *http.Request, apiGwRequest events.APIGatewayV2HTTPRequest) (*http.Request, error) {
	stageVars, err := json.Marshal(apiGwRequest.StageVariables)
	if err != nil {
		log.Println("Could not marshal stage variables for custom header")
		return nil, err
	}
	req.Header.Add(APIGwStageVarsHeader, string(stageVars))
	apiGwContext, err := json.Marshal(apiGwRequest.RequestContext)
	if err != nil {
		log.Println("Could not Marshal API GW context for custom header")
		return req, err
	}
	req.Header.Add(APIGwContextHeader, string(apiGwContext))
	return req, nil
}

func addToContextV2(ctx context.Context, req *http.Request, apiGwRequest events.APIGatewayV2HTTPRequest) *http.Request {
	lc, _ := lambdacontext.FromContext(ctx)
	rc := requestContextV2{lambdaContext: lc, gatewayProxyContext: apiGwRequest.RequestContext, stageVars: apiGwRequest.StageVariables}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

// GetAPIGatewayV2ContextFromContext retrieve APIGatewayProxyRequestContext from context.Context
func GetAPIGatewayV2ContextFromContext(ctx context.Context) (events.APIGatewayV2HTTPRequestContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextV2)
	return v.gatewayProxyContext, ok
}

// GetRuntimeContextFromContextV2 retrieve Lambda Runtime Context from context.Context
func GetRuntimeContextFromContextV2(ctx context.Context) (*lambdacontext.LambdaContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextV2)
	return v.lambdaContext, ok
}

// GetStageVarsFromContextV2 retrieve stage variables from context
func GetStageVarsFromContextV2(ctx context.Context) (map[string]string, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextV2)
	return v.stageVars, ok
}

type requestContextV2 struct {
	lambdaContext       *lambdacontext.LambdaContext
	gatewayProxyContext events.APIGatewayV2HTTPRequestContext
	stageVars           map[string]string
}
