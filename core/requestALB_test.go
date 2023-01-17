package core_test

import (
	"context"
	"encoding/base64"
	"math/rand"
	"strings"

	"github.com/awslabs/aws-lambda-go-api-proxy/core"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequestAccessorALB tests", func() {
	Context("ALB event conversion", func() {
		accessor := core.RequestAccessorALB{}
		qs := make(map[string]string)
		mvh := make(map[string][]string)
		mvqs := make(map[string][]string)
		hdr := make(map[string]string)
		qs["UniqueId"] = "12345"
		mvh["accept"] = []string{"test", "one"}
		mvh["connection"] = []string{"keep-alive"}
		mvh["host"] = []string{"lambda-test-alb-1234567.us-east-1.elb.amazonaws.com"}
		hdr["header1"] = "Testhdr1"
		hdr["header2"] = "Testhdr2"
		//multivalue querystrings
		mvqs["k1"] = []string{"t1"}
		mvqs["k2"] = []string{"t2"}
		bdy := "Test BODY"
		basicRequest := getALBProxyRequest("/hello", "GET", getALBRequestContext(), false, hdr, bdy, qs, mvh, nil)

		It("Correctly converts a basic event", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basicRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("/hello?UniqueId=12345").To(Equal(httpReq.RequestURI))
			Expect("GET").To(Equal(httpReq.Method))
			headers := basicRequest.Headers
			Expect(2).To(Equal(len(headers)))
			mvhs := basicRequest.MultiValueHeaders
			Expect(3).To(Equal(len(mvhs)))
			mvqs := basicRequest.MultiValueQueryStringParameters
			Expect(0).To(Equal(len(mvqs)))

		})

		binaryBody := make([]byte, 256)
		_, err := rand.Read(binaryBody)
		if err != nil {
			Fail("Could not generate random binary body")
		}

		encodedBody := base64.StdEncoding.EncodeToString(binaryBody)

		binaryRequest := getALBProxyRequest("/hello", "POST", getALBRequestContext(), true, hdr, bdy, qs, mvh, nil)
		binaryRequest.Body = encodedBody
		binaryRequest.IsBase64Encoded = true

		It("Decodes a base64 encoded body", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), binaryRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("/hello?UniqueId=12345").To(Equal(httpReq.RequestURI))
			Expect("POST").To(Equal(httpReq.Method))

			Expect(err).To(BeNil())

		})

		mqsRequest := getALBProxyRequest("/hello", "GET", getALBRequestContext(), false, hdr, bdy, qs, mvh, nil)
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
			Expect("1").To(Equal(query["hello"][0]))
			Expect("2").To(Equal(query["world"][0]))

		})

		qsRequest := getALBProxyRequest("/hello", "GET", getALBRequestContext(), false, hdr, bdy, qs, mvh, nil)
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

		// If multivaluehaders are set then it only passes the multivalue headers to the http.Request
		mvhRequest := getALBProxyRequest("/hello", "GET", getALBRequestContext(), false, hdr, bdy, qs, nil, mvqs)
		mvhRequest.MultiValueHeaders = map[string][]string{
			"accept":     {"test", "one"},
			"connection": {"keep-alive"},
			"host":       {"lambda-test-alb-1234567.us-east-1.elb.amazonaws.com"},
		}
		It("Populates multiple value headers correctly", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), mvhRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))

			headers := httpReq.Header
			Expect(3).To(Equal(len(headers)))

			for k, value := range headers {
				Expect(value).To(Equal(mvhRequest.MultiValueHeaders[strings.ToLower(k)]))
			}

		})
		// If multivaluehaders are set then it only passes the multivalue headers to the http.Request
		svhRequest := getALBProxyRequest("/hello", "GET", getALBRequestContext(), false, hdr, bdy, qs, mvh, mvqs)
		svhRequest.Headers = map[string]string{
			"header1": "Testhdr1",
			"header2": "Testhdr2"}

		It("Populates single value headers correctly", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), svhRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))

			headers := httpReq.Header
			Expect(3).To(Equal(len(headers)))

			for k, value := range headers {
				Expect(value).To(Equal(mvhRequest.MultiValueHeaders[strings.ToLower(k)]))
			}

		})

		basePathRequest := getALBProxyRequest("/app1/orders", "GET", getALBRequestContext(), false, hdr, bdy, qs, mvh, nil)

		It("Stips the base path correct", func() {
			accessor.StripBasePath("app1")
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basePathRequest)

			Expect(err).To(BeNil())
			Expect("/orders").To(Equal(httpReq.URL.Path))
			Expect("/orders?UniqueId=12345").To(Equal(httpReq.RequestURI))
		})

		contextRequest := getALBProxyRequest("orders", "GET", getALBRequestContext(), false, hdr, bdy, qs, mvh, mvqs)
		contextRequest.RequestContext = getALBRequestContext()

		It("Populates context header correctly", func() {
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())
			Expect(4).To(Equal(len(httpReq.Header)))
			Expect(httpReq.Header.Get(core.ALBContextHeader)).ToNot(BeNil())
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

	Context("Retrieves ALB Target Group Request context", func() {
		It("Returns a correctly unmarshalled object", func() {
			qs := make(map[string]string)
			mvh := make(map[string][]string)
			hdr := make(map[string]string)
			mvqs := make(map[string][]string)
			qs["UniqueId"] = "12345"
			mvh["accept"] = []string{"*/*", "/"}
			mvh["connection"] = []string{"keep-alive"}
			mvh["host"] = []string{"lambda-test-alb-1234567.us-east-1.elb.amazonaws.com"}
			mvqs["key1"] = []string{"Test1"}
			mvqs["key2"] = []string{"test2"}
			hdr["header1"] = "Testhdr1"
			bdy := "Test BODY2"

			contextRequest := getALBProxyRequest("/orders", "GET", getALBRequestContext(), false, hdr, bdy, qs, mvh, mvqs)
			contextRequest.RequestContext = getALBRequestContext()

			accessor := core.RequestAccessorALB{}
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())

			headerContext, err := accessor.GetContextALB(httpReq)
			Expect(err).To(BeNil())
			Expect(headerContext).ToNot(BeNil())
			Expect("arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/lambda-target/abcdefgh").To(Equal(headerContext.ELB.TargetGroupArn))
			proxyContext, ok := core.GetTargetGroupRequetFromContextALB(httpReq.Context())
			// should fail because using header proxy method
			Expect(ok).To(BeFalse())

			httpReq, err = accessor.EventToRequestWithContext(context.Background(), contextRequest)
			Expect(err).To(BeNil())
			proxyContext, ok = core.GetTargetGroupRequetFromContextALB(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect("arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/lambda-target/abcdefgh").To(Equal(proxyContext.ELB.TargetGroupArn))
			runtimeContext, ok := core.GetRuntimeContextFromContextALB(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(runtimeContext).To(BeNil())

			lambdaContext := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{AwsRequestID: "abc123"})
			httpReq, err = accessor.EventToRequestWithContext(lambdaContext, contextRequest)
			Expect(err).To(BeNil())

			headerContext, err = accessor.GetContextALB(httpReq)
			// should fail as new context method doesn't populate headers
			Expect(err).ToNot(BeNil())
			proxyContext, ok = core.GetTargetGroupRequetFromContextALB(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect("arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/lambda-target/abcdefgh").To(Equal(proxyContext.ELB.TargetGroupArn))
			runtimeContext, ok = core.GetRuntimeContextFromContextALB(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(runtimeContext).ToNot(BeNil())

		})
	})
})

func getALBProxyRequest(path string, method string, requestCtx events.ALBTargetGroupRequestContext,
	is64 bool, header map[string]string, body string, qs map[string]string, mvh map[string][]string, mvqsp map[string][]string) events.ALBTargetGroupRequest {
	return events.ALBTargetGroupRequest{
		HTTPMethod:                      method,
		Path:                            path,
		QueryStringParameters:           qs,
		MultiValueQueryStringParameters: mvqsp,
		Headers:                         header,
		MultiValueHeaders:               mvh,
		RequestContext:                  requestCtx,
		IsBase64Encoded:                 is64,
		Body:                            body,
	}
}

func getALBRequestContext() events.ALBTargetGroupRequestContext {
	return events.ALBTargetGroupRequestContext{
		ELB: events.ELBContext{
			TargetGroupArn: "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/lambda-target/abcdefgh",
		},
	}
}
