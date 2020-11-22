package core_test

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequestAccessorV2 tests", func() {
	Context("event conversion", func() {
		accessor := core.RequestAccessorV2{}
		basicRequest := getProxyRequestV2("/hello", "GET")
		It("Correctly converts a basic event", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basicRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("/hello").To(Equal(httpReq.RequestURI))
			Expect("GET").To(Equal(httpReq.Method))
		})

		basicRequest = getProxyRequestV2("/hello", "get")
		It("Converts method to uppercase", func() {
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(basicRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("/hello").To(Equal(httpReq.RequestURI))
			Expect("GET").To(Equal(httpReq.Method))
		})

		binaryBody := make([]byte, 256)
		_, err := rand.Read(binaryBody)
		if err != nil {
			Fail("Could not generate random binary body")
		}

		encodedBody := base64.StdEncoding.EncodeToString(binaryBody)

		binaryRequest := getProxyRequestV2("/hello", "POST")
		binaryRequest.Body = encodedBody
		binaryRequest.IsBase64Encoded = true

		It("Decodes a base64 encoded body", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), binaryRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("/hello").To(Equal(httpReq.RequestURI))
			Expect("POST").To(Equal(httpReq.Method))

			bodyBytes, err := ioutil.ReadAll(httpReq.Body)

			Expect(err).To(BeNil())
			Expect(len(binaryBody)).To(Equal(len(bodyBytes)))
			Expect(binaryBody).To(Equal(bodyBytes))
		})

		mqsRequest := getProxyRequestV2("/hello", "GET")
		mqsRequest.RawQueryString = "hello=1&world=2&world=3"
		mqsRequest.QueryStringParameters = map[string]string{
			"hello": "1",
			"world": "2",
		}
		It("Populates multiple value query string correctly", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), mqsRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect(httpReq.RequestURI).To(ContainSubstring("hello=1"))
			Expect(httpReq.RequestURI).To(ContainSubstring("world=2"))
			Expect(httpReq.RequestURI).To(ContainSubstring("world=3"))
			Expect("GET").To(Equal(httpReq.Method))

			query := httpReq.URL.Query()
			Expect(2).To(Equal(len(query)))
			Expect(query["hello"]).ToNot(BeNil())
			Expect(query["world"]).ToNot(BeNil())
			Expect(1).To(Equal(len(query["hello"])))
			Expect(2).To(Equal(len(query["world"])))
			Expect("1").To(Equal(query["hello"][0]))
			Expect("2").To(Equal(query["world"][0]))
			Expect("3").To(Equal(query["world"][1]))
		})

		// Support `QueryStringParameters` for backward compatibility.
		// https://github.com/awslabs/aws-lambda-go-api-proxy/issues/37
		qsRequest := getProxyRequestV2("/hello", "GET")
		qsRequest.QueryStringParameters = map[string]string{
			"hello": "1",
			"world": "2",
		}
		It("Populates query string correctly", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), qsRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect(httpReq.RequestURI).To(ContainSubstring("hello=1"))
			Expect(httpReq.RequestURI).To(ContainSubstring("world=2"))
			Expect("GET").To(Equal(httpReq.Method))

			query := httpReq.URL.Query()
			Expect(2).To(Equal(len(query)))
			Expect(query["hello"]).ToNot(BeNil())
			Expect(query["world"]).ToNot(BeNil())
			Expect(1).To(Equal(len(query["hello"])))
			Expect(1).To(Equal(len(query["world"])))
			Expect("1").To(Equal(query["hello"][0]))
			Expect("2").To(Equal(query["world"][0]))
		})

		mvhRequest := getProxyRequestV2("/hello", "GET")
		mvhRequest.Headers = map[string]string{
			"hello": "1",
			"world": "2,3",
		}

		It("Populates multiple value headers correctly", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), mvhRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))

			headers := httpReq.Header
			Expect(2).To(Equal(len(headers)))

			for k, value := range headers {
				Expect(strings.Join(value, ",")).To(Equal(mvhRequest.Headers[strings.ToLower(k)]))
			}
		})

		svhRequest := getProxyRequestV2("/hello", "GET")
		svhRequest.Headers = map[string]string{
			"hello": "1",
			"world": "2",
		}
		It("Populates single value headers correctly", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), svhRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))

			headers := httpReq.Header
			Expect(2).To(Equal(len(headers)))

			for k, value := range headers {
				Expect(value[0]).To(Equal(svhRequest.Headers[strings.ToLower(k)]))
			}
		})

		basePathRequest := getProxyRequestV2("/app1/orders", "GET")

		It("Stips the base path correct", func() {
			accessor.StripBasePath("app1")
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basePathRequest)

			Expect(err).To(BeNil())
			Expect("/orders").To(Equal(httpReq.URL.Path))
			Expect("/orders").To(Equal(httpReq.RequestURI))
		})

		contextRequest := getProxyRequestV2("orders", "GET")
		contextRequest.RequestContext = getRequestContextV2()

		It("Populates context header correctly", func() {
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())
			Expect(2).To(Equal(len(httpReq.Header)))
			Expect(httpReq.Header.Get(core.APIGwContextHeader)).ToNot(BeNil())
		})
	})

	Context("StripBasePath tests", func() {
		accessor := core.RequestAccessorV2{}
		It("Adds prefix slash", func() {
			basePath := accessor.StripBasePath("app1")
			Expect("/app1").To(Equal(basePath))
		})

		It("Removes trailing slash", func() {
			basePath := accessor.StripBasePath("/app1/")
			Expect("/app1").To(Equal(basePath))
		})

		It("Ignores blank strings", func() {
			basePath := accessor.StripBasePath("  ")
			Expect("").To(Equal(basePath))
		})
	})

	Context("Retrieves API Gateway context", func() {
		It("Returns a correctly unmarshalled object", func() {
			contextRequest := getProxyRequestV2("orders", "GET")
			contextRequest.RequestContext = getRequestContextV2()

			accessor := core.RequestAccessorV2{}
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())

			headerContext, err := accessor.GetAPIGatewayContextV2(httpReq)
			Expect(err).To(BeNil())
			Expect(headerContext).ToNot(BeNil())
			Expect("x").To(Equal(headerContext.AccountID))
			Expect("x").To(Equal(headerContext.RequestID))
			Expect("x").To(Equal(headerContext.APIID))
			proxyContext, ok := core.GetAPIGatewayV2ContextFromContext(httpReq.Context())
			// should fail because using header proxy method
			Expect(ok).To(BeFalse())

			httpReq, err = accessor.EventToRequestWithContext(context.Background(), contextRequest)
			Expect(err).To(BeNil())
			proxyContext, ok = core.GetAPIGatewayV2ContextFromContext(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect("x").To(Equal(proxyContext.APIID))
			Expect("x").To(Equal(proxyContext.RequestID))
			Expect("x").To(Equal(proxyContext.APIID))
			Expect("prod").To(Equal(proxyContext.Stage))
			runtimeContext, ok := core.GetRuntimeContextFromContextV2(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(runtimeContext).To(BeNil())

			lambdaContext := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{AwsRequestID: "abc123"})
			httpReq, err = accessor.EventToRequestWithContext(lambdaContext, contextRequest)
			Expect(err).To(BeNil())

			headerContext, err = accessor.GetAPIGatewayContextV2(httpReq)
			// should fail as new context method doesn't populate headers
			Expect(err).ToNot(BeNil())
			proxyContext, ok = core.GetAPIGatewayV2ContextFromContext(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect("x").To(Equal(proxyContext.APIID))
			Expect("x").To(Equal(proxyContext.RequestID))
			Expect("x").To(Equal(proxyContext.APIID))
			Expect("prod").To(Equal(proxyContext.Stage))
			runtimeContext, ok = core.GetRuntimeContextFromContextV2(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(runtimeContext).ToNot(BeNil())
			Expect("abc123").To(Equal(runtimeContext.AwsRequestID))
		})

		It("Populates stage variables correctly", func() {
			varsRequest := getProxyRequest("orders", "GET")
			varsRequest.StageVariables = getStageVariables()

			accessor := core.RequestAccessor{}
			httpReq, err := accessor.ProxyEventToHTTPRequest(varsRequest)
			Expect(err).To(BeNil())

			stageVars, err := accessor.GetAPIGatewayStageVars(httpReq)
			Expect(err).To(BeNil())
			Expect(2).To(Equal(len(stageVars)))
			Expect(stageVars["var1"]).ToNot(BeNil())
			Expect(stageVars["var2"]).ToNot(BeNil())
			Expect("value1").To(Equal(stageVars["var1"]))
			Expect("value2").To(Equal(stageVars["var2"]))

			stageVars, ok := core.GetStageVarsFromContext(httpReq.Context())
			// not present in context
			Expect(ok).To(BeFalse())

			httpReq, err = accessor.EventToRequestWithContext(context.Background(), varsRequest)
			Expect(err).To(BeNil())

			stageVars, err = accessor.GetAPIGatewayStageVars(httpReq)
			// should not be in headers
			Expect(err).ToNot(BeNil())

			stageVars, ok = core.GetStageVarsFromContext(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(2).To(Equal(len(stageVars)))
			Expect(stageVars["var1"]).ToNot(BeNil())
			Expect(stageVars["var2"]).ToNot(BeNil())
			Expect("value1").To(Equal(stageVars["var1"]))
			Expect("value2").To(Equal(stageVars["var2"]))
		})

		It("Populates the default hostname correctly", func() {

			basicRequest := getProxyRequest("orders", "GET")
			basicRequest.RequestContext = getRequestContext()
			accessor := core.RequestAccessor{}
			httpReq, err := accessor.ProxyEventToHTTPRequest(basicRequest)
			Expect(err).To(BeNil())

			Expect(basicRequest.RequestContext.DomainName).To(Equal(httpReq.Host))
			Expect(basicRequest.RequestContext.DomainName).To(Equal(httpReq.URL.Host))
		})

		It("Uses a custom hostname", func() {
			myCustomHost := "http://my-custom-host.com"
			os.Setenv(core.CustomHostVariable, myCustomHost)
			basicRequest := getProxyRequest("orders", "GET")
			accessor := core.RequestAccessor{}
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basicRequest)
			Expect(err).To(BeNil())

			Expect(myCustomHost).To(Equal("http://" + httpReq.Host))
			Expect(myCustomHost).To(Equal("http://" + httpReq.URL.Host))
			os.Unsetenv(core.CustomHostVariable)
		})

		It("Strips terminating / from hostname", func() {
			myCustomHost := "http://my-custom-host.com"
			os.Setenv(core.CustomHostVariable, myCustomHost+"/")
			basicRequest := getProxyRequest("orders", "GET")
			accessor := core.RequestAccessor{}
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basicRequest)
			Expect(err).To(BeNil())

			Expect(myCustomHost).To(Equal("http://" + httpReq.Host))
			Expect(myCustomHost).To(Equal("http://" + httpReq.URL.Host))
			os.Unsetenv(core.CustomHostVariable)
		})
	})
})

func getProxyRequestV2(path string, method string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Path:   path,
				Method: method,
			},
		},
		RawPath: path,
	}
}

func getRequestContextV2() events.APIGatewayV2HTTPRequestContext {
	return events.APIGatewayV2HTTPRequestContext{
		AccountID:  "x",
		RequestID:  "x",
		APIID:      "x",
		Stage:      "prod",
		DomainName: "12abcdefgh.execute-api.us-east-2.amazonaws.com",
	}
}
