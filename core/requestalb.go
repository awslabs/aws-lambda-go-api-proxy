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
	// ALBTgContextHeader is the custom header key used to store the
	// ALB Target Group Request context. To access the Context properties use the
	// RequestAccessorALB method of the RequestAccessorALB object.
	ALBTgContextHeader = "X-Golambdaproxy-Albtargetgroup-Context"
)

// RequestAccessorALB objects give access to custom API Gateway properties
// in the request.
type RequestAccessorALB struct{}

// GetALBTargetGroupRequestContext extracts the ALB Target Group Responce context object from a
// request's custom header.
// Returns a populated events.ALBTargetGroupRequestContext object from
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

// ProxyEventToHTTPRequest converts an ALB Target Group proxy event into a http.Request object.
// Returns the populated http request with additional two custom headers for ALB Tg Req context.
// To access these properties use the GetALBTargetGroupRequestContext method of the RequestAccessor object.
func (r *RequestAccessorALB) ProxyEventToHTTPRequest(req events.ALBTargetGroupRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeaderALB(httpRequest, req)
}

// EventToRequestWithContext converts an ALB Target Group proxy event and context into an http.Request object.
// Returns the populated http request with lambda context, ALBTargetGroupRequestContext as part of its context.
// Access those using GetRuntimeContextFromContextALB and GetRuntimeContextFromContext functions in this package.
func (r *RequestAccessorALB) EventToRequestWithContext(ctx context.Context, req events.ALBTargetGroupRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContextALB(ctx, httpRequest, req), nil
}

// EventToRequest converts an ALB Target group proxy event into an http.Request object.
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

	for headerKey, headerValue := range req.MultiValueHeaders {
		for _, arrVal := range headerValue {
			for _, val := range strings.Split(arrVal, ",") {
				httpRequest.Header.Add(headerKey, strings.Trim(val, " "))
			}
		}
	}

	httpRequest.RequestURI = httpRequest.URL.RequestURI()

	return httpRequest, nil
}

func addToHeaderALB(req *http.Request, albTgRequest events.ALBTargetGroupRequest) (*http.Request, error) {
	//req.Header.Add(ALBTgContextHeader, albTgRequest.RequestContext.ELB.TargetGroupArn)
	albTgContext, err := json.Marshal(albTgRequest.RequestContext)
	if err != nil {
		log.Println("Could not Marshal ALB Tg context for custom header")
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

// GetRuntimeContextFromContextALB retrieve Lambda Runtime Context from context.Context
func GetRuntimeContextFromContextALB(ctx context.Context) (*lambdacontext.LambdaContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContextALB)
	return v.lambdaContext, ok
}

type requestContextALB struct {
	lambdaContext       *lambdacontext.LambdaContext
	gatewayProxyContext events.ALBTargetGroupRequestContext
}
