// Package core provides utility methods that help convert ALB events
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
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

const (
	// FnURLContextHeader is the custom header key used to store the
	// Function URL context. To access the Context properties use the
	// GetContext method of the RequestAccessorFnURL object.
	FnURLContextHeader = "X-GoLambdaProxy-FnURL-Context"
)

// RequestAccessorFnURL objects give access to custom Function URL properties
// in the request.
type RequestAccessorFnURL struct {
	stripBasePath string
}

// GetALBContext extracts the ALB context object from a request's custom header.
// Returns a populated events.ALBTargetGroupRequestContext object from the request.
func (r *RequestAccessorFnURL) GetContext(req *http.Request) (events.LambdaFunctionURLRequestContext, error) {
	if req.Header.Get(FnURLContextHeader) == "" {
		return events.LambdaFunctionURLRequestContext{}, errors.New("no context header in request")
	}
	context := events.LambdaFunctionURLRequestContext{}
	err := json.Unmarshal([]byte(req.Header.Get(FnURLContextHeader)), &context)
	if err != nil {
		log.Println("Error while unmarshalling context")
		log.Println(err)
		return events.LambdaFunctionURLRequestContext{}, err
	}
	return context, nil
}

// StripBasePath instructs the RequestAccessor object that the given base
// path should be removed from the request path before sending it to the
// framework for routing. This is used when API Gateway is configured with
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

// FunctionURLEventToHTTPRequest converts an a Function URL event into a http.Request object.
// Returns the populated http request with additional custom header for the Function URL context.
// To access these properties use the GetContext method of the RequestAccessorFnURL object.
func (r *RequestAccessorFnURL) FunctionURLEventToHTTPRequest(req events.LambdaFunctionURLRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeaderFnURL(httpRequest, req)
}

// FunctionURLEventToHTTPRequestWithContext converts a Function URL event and context into an http.Request object.
// Returns the populated http request with lambda context, Function URL RequestContext as part of its context.
func (r *RequestAccessorFnURL) FunctionURLEventToHTTPRequestWithContext(ctx context.Context, req events.LambdaFunctionURLRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContextFnURL(ctx, httpRequest, req), nil
}

// EventToRequest converts a Function URL event into an http.Request object.
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

	path = serverAddress + path + "?" + req.RawQueryString

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.RequestContext.HTTP.Method),
		path,
		bytes.NewReader(decodedBody),
	)

	if err != nil {
		fmt.Printf("Could not convert request %s:%s to http.Request\n", req.RequestContext.HTTP.Method, req.RawPath)
		log.Println(err)
		return nil, err
	}

	for header, val := range req.Headers {
		httpRequest.Header.Add(header, val)
	}

	httpRequest.RemoteAddr = req.RequestContext.HTTP.SourceIP
	httpRequest.RequestURI = httpRequest.URL.RequestURI()

	return httpRequest, nil
}

func addToHeaderFnURL(req *http.Request, fnUrlRequest events.LambdaFunctionURLRequest) (*http.Request, error) {
	ctx, err := json.Marshal(fnUrlRequest.RequestContext)
	if err != nil {
		log.Println("Could not Marshal Function URL context for custom header")
		return req, err
	}
	req.Header.Set(FnURLContextHeader, string(ctx))
	return req, nil
}

// adds context data to http request so we can pass
func addToContextFnURL(ctx context.Context, req *http.Request, fnUrlRequest events.LambdaFunctionURLRequest) *http.Request {
	lc, _ := lambdacontext.FromContext(ctx)
	rc := requestContextFnURL{lambdaContext: lc, fnUrlContext: fnUrlRequest.RequestContext}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

type requestContextFnURL struct {
	lambdaContext *lambdacontext.LambdaContext
	fnUrlContext  events.LambdaFunctionURLRequestContext
}
