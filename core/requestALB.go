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
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

const (
	// ALBContextHeader is the custom header key used to store the
	// ALB ELB context. To access the Context properties use the
	// GetALBContext method of the RequestAccessorALB object.
	ALBContextHeader = "X-GoLambdaProxy-ALB-Context"
)

// RequestAccessorALB objects give access to custom ALB Target Group properties
// in the request.
type RequestAccessorALB struct {
	stripBasePath string
}

// GetALBContext extracts the ALB context object from a request's custom header.
// Returns a populated events.ALBTargetGroupRequestContext object from the request.
func (r *RequestAccessorALB) GetContextALB(req *http.Request) (events.ALBTargetGroupRequestContext, error) {
	if req.Header.Get(ALBContextHeader) == "" {
		return events.ALBTargetGroupRequestContext{}, errors.New("no context header in request")
	}
	context := events.ALBTargetGroupRequestContext{}
	err := json.Unmarshal([]byte(req.Header.Get(ALBContextHeader)), &context)
	if err != nil {
		log.Println("Error while unmarshalling context")
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

// ProxyEventToHTTPRequest converts an ALB Target Group Request event into a http.Request object.
// Returns the populated http request with additional custom header for the ALB context.
// To access these properties use the GetALBContext method of the RequestAccessorALB object.
func (r *RequestAccessorALB) ProxyEventToHTTPRequest(req events.ALBTargetGroupRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeaderALB(httpRequest, req)
}

// EventToRequestWithContext converts an ALB Target Group Request event and context into an http.Request object.
// Returns the populated http request with lambda context, ALB TargetGroup RequestContext as part of its context.
func (r *RequestAccessorALB) EventToRequestWithContext(ctx context.Context, req events.ALBTargetGroupRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContextALB(ctx, httpRequest, req), nil
}

// EventToRequest converts an ALB TargetGroup event into an http.Request object.
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
	serverAddress := "https://" + req.Headers["host"]
	//  if customAddress, ok := os.LookupEnv(CustomHostVariable); ok {
	//  	serverAddress = customAddress
	//  }
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

func addToHeaderALB(req *http.Request, albRequest events.ALBTargetGroupRequest) (*http.Request, error) {
	albContext, err := json.Marshal(albRequest.RequestContext)
	if err != nil {
		log.Println("Could not Marshal ALB context for custom header")
		return req, err
	}
	req.Header.Set(ALBContextHeader, string(albContext))
	return req, nil
}

// adds context data to http request so we can pass
func addToContextALB(ctx context.Context, req *http.Request, albRequest events.ALBTargetGroupRequest) *http.Request {
	lc, _ := lambdacontext.FromContext(ctx)
	rc := requestContextALB{lambdaContext: lc, albContext: albRequest.RequestContext}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

// GetALBTargetGroupRequestFromContext retrieve ALBTargetGroupt from context.Context
func GetTargetGroupRequetFromContextALB(ctx context.Context) (events.ALBTargetGroupRequestContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextALB)
	return v.albContext, ok
}

// GetRuntimeContextFromContext retrieve Lambda Runtime Context from context.Context
func GetRuntimeContextFromContextALB(ctx context.Context) (*lambdacontext.LambdaContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextALB)
	return v.lambdaContext, ok
}

type requestContextALB struct {
	lambdaContext *lambdacontext.LambdaContext
	albContext    events.ALBTargetGroupRequestContext
}
