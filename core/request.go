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

// CustomHostVariable is the name of the environment variable that contains
// the custom hostname for the request. If this variable is not set the framework
// reverts to `RequestContext.DomainName`. The value for a custom host should
// include a protocol: http://my-custom.host.com
const CustomHostVariable = "GO_API_HOST"

// APIGwContextHeader is the custom header key used to store the
// API Gateway context. To access the Context properties use the
// GetAPIGatewayContext method of the RequestAccessor object.
const APIGwContextHeader = "X-GoLambdaProxy-ApiGw-Context"

// APIGwStageVarsHeader is the custom header key used to store the
// API Gateway stage variables. To access the stage variable values
// use the GetAPIGatewayStageVars method of the RequestAccessor object.
const APIGwStageVarsHeader = "X-GoLambdaProxy-ApiGw-StageVars"

// RequestAccessor objects give access to custom API Gateway properties
// in the request.
type RequestAccessor struct {
	stripBasePath string
}

// GetAPIGatewayContext extracts the API Gateway context object from a
// request's custom header.
// Returns a populated events.APIGatewayProxyRequestContext object from
// the request.
func (r *RequestAccessor) GetAPIGatewayContext(req *http.Request) (events.APIGatewayProxyRequestContext, error) {
	if req.Header.Get(APIGwContextHeader) == "" {
		return events.APIGatewayProxyRequestContext{}, errors.New("No context header in request")
	}
	context := events.APIGatewayProxyRequestContext{}
	err := json.Unmarshal([]byte(req.Header.Get(APIGwContextHeader)), &context)
	if err != nil {
		log.Println("Erorr while unmarshalling context")
		log.Println(err)
		return events.APIGatewayProxyRequestContext{}, err
	}
	return context, nil
}

// GetAPIGatewayStageVars extracts the API Gateway stage variables from a
// request's custom header.
// Returns a map[string]string of the stage variables and their values from
// the request.
func (r *RequestAccessor) GetAPIGatewayStageVars(req *http.Request) (map[string]string, error) {
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
func (r *RequestAccessor) StripBasePath(basePath string) string {
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
func (r *RequestAccessor) ProxyEventToHTTPRequest(req events.APIGatewayProxyRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeader(httpRequest, req)
}

// EventToRequestWithContext converts an API Gateway proxy event and context into an http.Request object.
// Returns the populated http request with lambda context, stage variables and APIGatewayProxyRequestContext as part of its context.
// Access those using GetAPIGatewayContextFromContext, GetStageVarsFromContext and GetRuntimeContextFromContext functions in this package.
func (r *RequestAccessor) EventToRequestWithContext(ctx context.Context, req events.APIGatewayProxyRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContext(ctx, httpRequest, req), nil
}

// EventToRequest converts an API Gateway proxy event into an http.Request object.
// Returns the populated request maintaining headers
func (r *RequestAccessor) EventToRequest(req events.APIGatewayProxyRequest) (*http.Request, error) {
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
	serverAddress := "https://" + req.RequestContext.DomainName
	if customAddress, ok := os.LookupEnv(CustomHostVariable); ok {
		serverAddress = customAddress
	}
	path = serverAddress + path

	if len(req.MultiValueQueryStringParameters) > 0 {
		queryString := ""
		for q, l := range req.MultiValueQueryStringParameters {
			for _, v := range l {
				if queryString != "" {
					queryString += "&"
				}
				queryString += url.QueryEscape(q) + "=" + url.QueryEscape(v)
			}
		}
		path += "?" + queryString
	} else if len(req.QueryStringParameters) > 0 {
		// Support `QueryStringParameters` for backward compatibility.
		// https://github.com/awslabs/aws-lambda-go-api-proxy/issues/37
		queryString := ""
		for q := range req.QueryStringParameters {
			if queryString != "" {
				queryString += "&"
			}
			queryString += url.QueryEscape(q) + "=" + url.QueryEscape(req.QueryStringParameters[q])
		}
		path += "?" + queryString
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

	if req.MultiValueHeaders != nil {
		for k, values := range req.MultiValueHeaders {
			for _, value := range values {
				httpRequest.Header.Add(k, value)
			}
		}
	} else {
		for h := range req.Headers {
			httpRequest.Header.Add(h, req.Headers[h])
		}
	}

	httpRequest.RequestURI = httpRequest.URL.RequestURI()

	return httpRequest, nil
}

func addToHeader(req *http.Request, apiGwRequest events.APIGatewayProxyRequest) (*http.Request, error) {
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

func addToContext(ctx context.Context, req *http.Request, apiGwRequest events.APIGatewayProxyRequest) *http.Request {
	lc, _ := lambdacontext.FromContext(ctx)
	rc := requestContext{lambdaContext: lc, gatewayProxyContext: apiGwRequest.RequestContext, stageVars: apiGwRequest.StageVariables}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

// GetAPIGatewayContextFromContext retrieve APIGatewayProxyRequestContext from context.Context
func GetAPIGatewayContextFromContext(ctx context.Context) (events.APIGatewayProxyRequestContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContext)
	return v.gatewayProxyContext, ok
}

// GetRuntimeContextFromContext retrieve Lambda Runtime Context from context.Context
func GetRuntimeContextFromContext(ctx context.Context) (*lambdacontext.LambdaContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContext)
	return v.lambdaContext, ok
}

// GetStageVarsFromContext retrieve stage variables from context
func GetStageVarsFromContext(ctx context.Context) (map[string]string, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContext)
	return v.stageVars, ok
}

type ctxKey struct{}

type requestContext struct {
	lambdaContext       *lambdacontext.LambdaContext
	gatewayProxyContext events.APIGatewayProxyRequestContext
	stageVars           map[string]string
}
