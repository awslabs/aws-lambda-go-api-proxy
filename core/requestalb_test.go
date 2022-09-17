package core_test

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"math/rand"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequestAccessorALB tests", func() {
	Context("event conversion", func() {
		accessor := core.RequestAccessorALB{}
		basicRequest := getProxyRequestALB("/hello", "GET")
		It("Correctly converts a basic event", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basicRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("/hello").To(Equal(httpReq.RequestURI))
			Expect("GET").To(Equal(httpReq.Method))
		})

		basicRequest = getProxyRequestALB("/hello", "get")
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

		binaryRequest := getProxyRequestALB("/hello", "POST")
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

		mqsRequest := getProxyRequestALB("/hello", "GET")
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

		// Support `QueryStringParameters` for backward compatibility.
		// https://github.com/awslabs/aws-lambda-go-api-proxy/issues/37
		qsRequest := getProxyRequestALB("/hello", "GET")
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

		mvhRequest := getProxyRequestALB("/hello", "GET")
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

		svhRequest := getProxyRequestALB("/hello", "GET")
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

		basePathRequest := getProxyRequestALB("/app1/orders", "GET")

		It("Stips the base path correct", func() {
			accessor.StripBasePath("app1")
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basePathRequest)

			Expect(err).To(BeNil())
			Expect("/orders").To(Equal(httpReq.URL.Path))
			Expect("/orders").To(Equal(httpReq.RequestURI))
		})

		contextRequest := getProxyRequestALB("orders", "GET")
		contextRequest.RequestContext = getRequestContextALB()

		It("Populates context header correctly", func() {
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())
			Expect(1).To(Equal(len(httpReq.Header)))
			Expect(httpReq.Header.Get(core.ALBTgContextHeader)).ToNot(BeNil())
		})
	})

	Context("StripBasePath tests", func() {
		accessor := core.RequestAccessorALB{}
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

	Context("Retrieves ALB Target Group context", func() {
		It("Returns a correctly unmarshalled object", func() {
			contextRequest := getProxyRequestALB("orders", "GET")
			contextRequest.RequestContext = getRequestContextALB()

			accessor := core.RequestAccessorALB{}
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())

			headerContext, err := accessor.GetALBTargetGroupRequestContext(httpReq)
			Expect(err).To(BeNil())
			Expect(headerContext).ToNot(BeNil())
			proxyContext, ok := core.GetALBTargetGroupContextFromContext(httpReq.Context())
			// should fail because using header proxy method
			Expect(ok).To(BeFalse())

			httpReq, err = accessor.EventToRequestWithContext(context.Background(), contextRequest)
			Expect(err).To(BeNil())
			proxyContext, ok = core.GetALBTargetGroupContextFromContext(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(proxyContext.ELB.TargetGroupArn).ToNot(BeNil())
			runtimeContext, ok := core.GetRuntimeContextFromContextALB(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(runtimeContext).To(BeNil())

			lambdaContext := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{AwsRequestID: "abc123"})
			httpReq, err = accessor.EventToRequestWithContext(lambdaContext, contextRequest)
			Expect(err).To(BeNil())

			headerContext, err = accessor.GetALBTargetGroupRequestContext(httpReq)
			// should fail as new context method doesn't populate headers
			Expect(err).ToNot(BeNil())
			proxyContext, ok = core.GetALBTargetGroupContextFromContext(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(proxyContext.ELB.TargetGroupArn).ToNot(BeNil())
			runtimeContext, ok = core.GetRuntimeContextFromContextALB(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(runtimeContext).ToNot(BeNil())
			Expect("abc123").To(Equal(runtimeContext.AwsRequestID))
		})

		It("Populates stage variables correctly", func() {
			varsRequest := getProxyRequestALB("orders", "GET")

			accessor := core.RequestAccessorALB{}
			httpReq, err := accessor.ProxyEventToHTTPRequest(varsRequest)
			Expect("/orders").To(Equal(httpReq.RequestURI))
			Expect(err).To(BeNil())

			httpReq, err = accessor.EventToRequestWithContext(context.Background(), varsRequest)
			Expect("/orders").To(Equal(httpReq.RequestURI))
			Expect(err).To(BeNil())
		})

		It("Populates the default hostname correctly", func() {

			basicRequest := getProxyRequestALB("orders", "GET")
			basicRequest.RequestContext = getRequestContextALB()
			accessor := core.RequestAccessorALB{}
			httpReq, err := accessor.ProxyEventToHTTPRequest(basicRequest)
			Expect("/orders").To(Equal(httpReq.RequestURI))
			Expect(err).To(BeNil())

			Expect(httpReq.RequestURI).To(ContainSubstring(basicRequest.Path))
		})
	})
})

func getProxyRequestALB(path string, method string) events.ALBTargetGroupRequest {
	return events.ALBTargetGroupRequest{
		RequestContext: events.ALBTargetGroupRequestContext{},
		Path:           path,
		HTTPMethod:     method,
		Headers:        map[string]string{},
	}
}

func getRequestContextALB() events.ALBTargetGroupRequestContext {
	return events.ALBTargetGroupRequestContext{
		ELB: events.ELBContext{
			TargetGroupArn: "foo",
		},
	}
}
