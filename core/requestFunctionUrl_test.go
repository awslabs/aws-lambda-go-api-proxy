package core_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequestAccessorFu tests", func() {
	Context("Function URL event conversion", func() {
		accessor := core.RequestAccessorFu{}
		qs := make(map[string]string)
		mvqs := make(map[string][]string)
		hdr := make(map[string]string)
		qs["UniqueId"] = "12345"
		hdr["header1"] = "Testhdr1"
		hdr["header2"] = "Testhdr2"
		// Multivalue query strings
		mvqs["k1"] = []string{"t1"}
		mvqs["k2"] = []string{"t2"}
		bdy := "Test BODY"
		basePathRequest := getFunctionUrlProxyRequest("/hello", getFunctionUrlRequestContext("/hello", "GET"), false, hdr, bdy, qs, mvqs)

		It("Correctly converts a basic event", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basePathRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("/hello?UniqueId=12345").To(Equal(httpReq.RequestURI))
			Expect("GET").To(Equal(httpReq.Method))
			headers := basePathRequest.Headers
			Expect(2).To(Equal(len(headers)))
		})

		binaryBody := make([]byte, 256)
		_, err := rand.Read(binaryBody)
		if err != nil {
			Fail("Could not generate random binary body")
		}

		encodedBody := base64.StdEncoding.EncodeToString(binaryBody)

		binaryRequest := getFunctionUrlProxyRequest("/hello", getFunctionUrlRequestContext("/hello", "POST"), true, hdr, bdy, qs, mvqs)
		binaryRequest.Body = encodedBody
		binaryRequest.IsBase64Encoded = true

		It("Decodes a base64 encoded body", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), binaryRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("/hello?UniqueId=12345").To(Equal(httpReq.RequestURI))
			Expect("POST").To(Equal(httpReq.Method))
		})

		mqsRequest := getFunctionUrlProxyRequest("/hello", getFunctionUrlRequestContext("/hello", "GET"), false, hdr, bdy, qs, mvqs)
		mqsRequest.RawQueryString = "hello=1&world=2&world=3"
		mqsRequest.QueryStringParameters = map[string]string{
			"hello": "1",
			"world": "2",
		}

		It("Populates query string correctly", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), mqsRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			fmt.Println("SDYFSDKFJDL")
			fmt.Printf("%v", httpReq.RequestURI)
			Expect(httpReq.RequestURI).To(ContainSubstring("hello=1"))
			Expect(httpReq.RequestURI).To(ContainSubstring("world=2"))
			Expect("GET").To(Equal(httpReq.Method))
			query := httpReq.URL.Query()
			Expect(2).To(Equal(len(query)))
			Expect(query["hello"]).ToNot(BeNil())
			Expect(query["world"]).ToNot(BeNil())
		})
	})

	Context("StripBasePath tests", func() {
		accessor := core.RequestAccessorFu{}
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
})

func getFunctionUrlProxyRequest(path string, requestCtx events.LambdaFunctionURLRequestContext,
	is64 bool, header map[string]string, body string, qs map[string]string, mvqs map[string][]string) events.LambdaFunctionURLRequest {
	return events.LambdaFunctionURLRequest{
		RequestContext:  requestCtx,
		RawPath:         path,
		RawQueryString:  generateQueryString(qs),
		Headers:         header,
		Body:            body,
		IsBase64Encoded: is64,
	}
}

func getFunctionUrlRequestContext(path, method string) events.LambdaFunctionURLRequestContext {
	return events.LambdaFunctionURLRequestContext{
		DomainName: "example.com",
		HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
			Method: method,
			Path:   path,
		},
	}
}

func generateQueryString(queryParameters map[string]string) string {
	var queryString string
	for key, value := range queryParameters {
		if queryString != "" {
			queryString += "&"
		}
		queryString += fmt.Sprintf("%s=%s", key, value)
	}
	return queryString
}
