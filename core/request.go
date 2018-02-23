// Package core provides utility methods that help convert proxy events
// into an http.Request and http.ResponseWriter
package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// DefaultServerAddress is prepended to the path of each incoming reuqest
const DefaultServerAddress = "https://aws-serverless-go-api.com"

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

// ProxyEventToHTTPRequest converts an API Gateway proxy event into an
// http.Request object.
// Returns the populated request with an additional two custom headers for the
// stage variables and API Gateway context. To access these properties use
// the GetAPIGatewayStageVars and GetAPIGatewayContext method of the RequestAccessor
// object.
func (r *RequestAccessor) ProxyEventToHTTPRequest(req events.APIGatewayProxyRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	queryString := ""
	if len(req.QueryStringParameters) > 0 {
		queryString = "?"
		queryCnt := 0
		for q := range req.QueryStringParameters {
			if queryCnt > 0 {
				queryString += "&"
			}
			queryString += url.QueryEscape(q) + "=" + url.QueryEscape(req.QueryStringParameters[q])
			queryCnt++
		}
	}

	path := req.Path
	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
		}
	}

	path = DefaultServerAddress + path

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.HTTPMethod),
		path+queryString,
		bytes.NewReader(decodedBody),
	)

	if err != nil {
		fmt.Printf("Could not convert request %s:%s to http.Request\n", req.HTTPMethod, req.Path)
		log.Println(err)
		return nil, err
	}

	for h := range req.Headers {
		httpRequest.Header.Add(h, req.Headers[h])
	}

	apiGwContext, err := json.Marshal(req.RequestContext)
	if err != nil {
		log.Println("Could not Marshal API GW context for custom header")
		return nil, err
	}
	stageVars, err := json.Marshal(req.StageVariables)
	if err != nil {
		log.Println("Could not marshal stage variables for custom header")
		return nil, err
	}
	httpRequest.Header.Add(APIGwContextHeader, string(apiGwContext))
	httpRequest.Header.Add(APIGwStageVarsHeader, string(stageVars))

	return httpRequest, nil
}
