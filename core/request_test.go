package core_test

import (
	"encoding/base64"
	"io/ioutil"
	"math/rand"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequestAccessor tests", func() {
	Context("event conversion", func() {
		accessor := core.RequestAccessor{}
		basicRequest := getProxyRequest("/hello", "GET")
		It("Correctly converts a basic event", func() {
			httpReq, err := accessor.ProxyEventToHTTPRequest(basicRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))
		})

		basicRequest = getProxyRequest("/hello", "get")
		It("Converts method to uppercase", func() {
			httpReq, err := accessor.ProxyEventToHTTPRequest(basicRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))
		})

		binaryBody := make([]byte, 256)
		_, err := rand.Read(binaryBody)
		if err != nil {
			Fail("Could not generate random binary body")
		}

		encodedBody := base64.StdEncoding.EncodeToString(binaryBody)

		binaryRequest := getProxyRequest("/hello", "POST")
		binaryRequest.Body = encodedBody
		binaryRequest.IsBase64Encoded = true

		It("Decodes a base64 encoded body", func() {
			httpReq, err := accessor.ProxyEventToHTTPRequest(binaryRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("POST").To(Equal(httpReq.Method))

			bodyBytes, err := ioutil.ReadAll(httpReq.Body)

			Expect(err).To(BeNil())
			Expect(len(binaryBody)).To(Equal(len(bodyBytes)))
			Expect(binaryBody).To(Equal(bodyBytes))
		})

		qsRequest := getProxyRequest("/hello", "GET")
		qsRequest.QueryStringParameters = map[string]string{
			"hello": "1",
			"world": "2",
		}
		It("Populates query string correctly", func() {
			httpReq, err := accessor.ProxyEventToHTTPRequest(qsRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
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

		basePathRequest := getProxyRequest("/app1/orders", "GET")

		It("Stips the base path correct", func() {
			accessor.StripBasePath("app1")
			httpReq, err := accessor.ProxyEventToHTTPRequest(basePathRequest)
			Expect(err).To(BeNil())
			Expect("/orders").To(Equal(httpReq.URL.Path))
		})

		contextRequest := getProxyRequest("orders", "GET")
		contextRequest.RequestContext = getRequestContext()

		It("Populates context header correctly", func() {
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())
			Expect(2).To(Equal(len(httpReq.Header)))
			Expect(httpReq.Header.Get(core.APIGwContextHeader)).ToNot(BeNil())
		})
	})

	Context("StripBasePath tests", func() {
		accessor := core.RequestAccessor{}
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
			contextRequest := getProxyRequest("orders", "GET")
			contextRequest.RequestContext = getRequestContext()

			accessor := core.RequestAccessor{}
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())

			context, err := accessor.GetAPIGatewayContext(httpReq)
			Expect(err).To(BeNil())
			Expect(context).ToNot(BeNil())
			Expect("x").To(Equal(context.AccountID))
			Expect("x").To(Equal(context.RequestID))
			Expect("x").To(Equal(context.APIID))
			Expect("prod").To(Equal(context.Stage))
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
		})
	})
})

func getProxyRequest(path string, method string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Path:       path,
		HTTPMethod: method,
	}
}

func getRequestContext() events.APIGatewayProxyRequestContext {
	return events.APIGatewayProxyRequestContext{
		AccountID: "x",
		RequestID: "x",
		APIID:     "x",
		Stage:     "prod",
	}
}

func getStageVariables() map[string]string {
	return map[string]string{
		"var1": "value1",
		"var2": "value2",
	}
}
