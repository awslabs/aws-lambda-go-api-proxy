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

const (
	// FuContextHeader is the custom header key used to store the
	// Function Url context. To access the Context properties use the
	// GetFunctionURLContext method of the RequestAccessorFu object.
	FnURLContextHeader = "X-GoLambdaProxy-Fu-Context"
)

// RequestAccessorFu objects give access to custom API Gateway properties
// in the request.
type RequestAccessorFnURL struct {
	stripBasePath string
}

// GetFunctionURLContext extracts the API Gateway context object from a
// request's custom header.
// Returns a populated events.LambdaFunctionURLRequestContext object from
// the request.
func (r *RequestAccessorFnURL) GetFunctionURLContext(req *http.Request) (events.LambdaFunctionURLRequestContext, error) {
	if req.Header.Get(APIGwContextHeader) == "" {
		return events.LambdaFunctionURLRequestContext{}, errors.New("No context header in request")
	}
	context := events.LambdaFunctionURLRequestContext{}
	err := json.Unmarshal([]byte(req.Header.Get(FnURLContextHeader)), &context)
	if err != nil {
		log.Println("Erorr while unmarshalling context")
		log.Println(err)
		return events.LambdaFunctionURLRequestContext{}, err
	}
	return context, nil
}

// StripBasePath instructs the RequestAccessor object that the given base
// path should be removed from the request path before sending it to the
// framework for routing. This is used when the Lambda is configured with
// base path mappings in custom domain names.
func (r *RequestAccessorFnURL) StripBasePath(basePath string) string {
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

// ProxyEventToHTTPRequest converts an Function URL proxy event into a http.Request object.
// Returns the populated http request with additional two custom headers for the stage variables and Function Url context.
// To access these properties use GetFunctionURLContext method of the RequestAccessor object.
func (r *RequestAccessorFnURL) ProxyEventToHTTPRequest(req events.LambdaFunctionURLRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeaderFunctionURL(httpRequest, req)
}

// EventToRequestWithContext converts an Function URL proxy event and context into an http.Request object.
// Returns the populated http request with lambda context, stage variables and APIGatewayProxyRequestContext as part of its context.
// Access those using GetFunctionURLContextFromContext and GetRuntimeContextFromContext functions in this package.
func (r *RequestAccessorFnURL) EventToRequestWithContext(ctx context.Context, req events.LambdaFunctionURLRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContextFunctionURL(ctx, httpRequest, req), nil
}

// EventToRequest converts an Function URL proxy event into an http.Request object.
// Returns the populated request maintaining headers
func (r *RequestAccessorFnURL) EventToRequest(req events.LambdaFunctionURLRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	path := req.RawPath
	// if RawPath empty is, populate from request context
	if len(path) == 0 {
		path = req.RequestContext.HTTP.Path
	}

	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
		}
		fmt.Printf("%v", path)
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

	httpRequest.RemoteAddr = req.RequestContext.HTTP.SourceIP

	for _, cookie := range req.Cookies {
		httpRequest.Header.Add("Cookie", cookie)
	}

	for headerKey, headerValue := range req.Headers {
		for _, val := range strings.Split(headerValue, ",") {
			httpRequest.Header.Add(headerKey, strings.Trim(val, " "))
		}
	}

	httpRequest.RequestURI = httpRequest.URL.RequestURI()

	return httpRequest, nil
}

func addToHeaderFunctionURL(req *http.Request, FunctionURLRequest events.LambdaFunctionURLRequest) (*http.Request, error) {
	apiGwContext, err := json.Marshal(FunctionURLRequest.RequestContext)
	if err != nil {
		log.Println("Could not Marshal API GW context for custom header")
		return req, err
	}
	req.Header.Add(APIGwContextHeader, string(apiGwContext))
	return req, nil
}

func addToContextFunctionURL(ctx context.Context, req *http.Request, FunctionURLRequest events.LambdaFunctionURLRequest) *http.Request {
	lc, _ := lambdacontext.FromContext(ctx)
	rc := requestContextFnURL{lambdaContext: lc, FunctionURLProxyContext: FunctionURLRequest.RequestContext}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

// GetFunctionURLContextFromContext retrieve APIGatewayProxyRequestContext from context.Context
func GetFunctionURLContextFromContext(ctx context.Context) (events.LambdaFunctionURLRequestContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextFnURL)
	return v.FunctionURLProxyContext, ok
}

// GetRuntimeContextFromContextFu retrieve Lambda Runtime Context from context.Context
func GetRuntimeContextFromContextFnURL(ctx context.Context) (*lambdacontext.LambdaContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextFnURL)
	return v.lambdaContext, ok
}

type requestContextFnURL struct {
	lambdaContext           *lambdacontext.LambdaContext
	FunctionURLProxyContext events.LambdaFunctionURLRequestContext
}
