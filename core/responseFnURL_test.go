package core

import (
	"math/rand"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FunctionURLResponseWriter tests", func() {
	Context("writing to response object", func() {
		response := NewFunctionURLResponseWriter()

		It("Sets the correct default status", func() {
			Expect(defaultStatusCode).To(Equal(response.status))
		})

		It("Initializes the headers map", func() {
			Expect(response.headers).ToNot(BeNil())
			Expect(0).To(Equal(len(response.headers)))
		})

		It("Writes headers correctly", func() {
			response.Header().Add("Content-Type", "application/json")

			Expect(1).To(Equal(len(response.headers)))
			Expect("application/json").To(Equal(response.headers["Content-Type"][0]))
		})

		It("Writes body content correctly", func() {
			binaryBody := make([]byte, 256)
			_, err := rand.Read(binaryBody)
			Expect(err).To(BeNil())

			written, err := response.Write(binaryBody)
			Expect(err).To(BeNil())
			Expect(len(binaryBody)).To(Equal(written))
		})

		It("Automatically set the status code to 200", func() {
			Expect(http.StatusOK).To(Equal(response.status))
		})

		It("Forces the status to a new code", func() {
			response.WriteHeader(http.StatusAccepted)
			Expect(http.StatusAccepted).To(Equal(response.status))
		})
	})

	Context("Automatically set response content type", func() {
		xmlBodyContent := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><note><to>Tove</to><from>Jani</from><heading>Reminder</heading><body>Don't forget me this weekend!</body></note>"
		htmlBodyContent := " <!DOCTYPE html><html><head><meta charset=\"UTF-8\"><title>Title of the document</title></head><body>Content of the document......</body></html>"

		It("Does not set the content type if it's already set", func() {
			resp := NewFunctionURLResponseWriter()
			resp.Header().Add("Content-Type", "application/json")

			resp.Write([]byte(xmlBodyContent))

			Expect("application/json").To(Equal(resp.Header().Get("Content-Type")))
			proxyResp, err := resp.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(1).To(Equal(len(proxyResp.Headers)))
			Expect("application/json").To(Equal(proxyResp.Headers["Content-Type"]))
			Expect(xmlBodyContent).To(Equal(proxyResp.Body))
		})

		It("Sets the content type to text/xml given the body", func() {
			resp := NewFunctionURLResponseWriter()
			resp.Write([]byte(xmlBodyContent))

			Expect("").ToNot(Equal(resp.Header().Get("Content-Type")))
			Expect(true).To(Equal(strings.HasPrefix(resp.Header().Get("Content-Type"), "text/xml;")))
			proxyResp, err := resp.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(1).To(Equal(len(proxyResp.Headers)))
			Expect(true).To(Equal(strings.HasPrefix(proxyResp.Headers["Content-Type"], "text/xml;")))
			Expect(xmlBodyContent).To(Equal(proxyResp.Body))
		})

		It("Sets the content type to text/html given the body", func() {
			resp := NewFunctionURLResponseWriter()
			resp.Write([]byte(htmlBodyContent))

			Expect("").ToNot(Equal(resp.Header().Get("Content-Type")))
			Expect(true).To(Equal(strings.HasPrefix(resp.Header().Get("Content-Type"), "text/html;")))
			proxyResp, err := resp.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(1).To(Equal(len(proxyResp.Headers)))
			Expect(true).To(Equal(strings.HasPrefix(proxyResp.Headers["Content-Type"], "text/html;")))
			Expect(htmlBodyContent).To(Equal(proxyResp.Body))
		})
	})

	Context("Export Lambda Function URL response", func() {
		emptyResponse := NewFunctionURLResponseWriter()
		emptyResponse.Header().Add("Content-Type", "application/json")

		It("Refuses empty responses with default status code", func() {
			_, err := emptyResponse.GetProxyResponse()
			Expect(err).ToNot(BeNil())
			Expect("Status code not set on response").To(Equal(err.Error()))
		})

		simpleResponse := NewFunctionURLResponseWriter()
		simpleResponse.Write([]byte("https://example.com"))
		simpleResponse.WriteHeader(http.StatusAccepted)

		It("Writes function URL response correctly", func() {
			FunctionURLResponse, err := simpleResponse.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(FunctionURLResponse).ToNot(BeNil())
			Expect(http.StatusAccepted).To(Equal(FunctionURLResponse.StatusCode))
		})
	})
})
