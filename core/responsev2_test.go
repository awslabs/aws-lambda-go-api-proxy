package core

import (
	"encoding/base64"
	"math/rand"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResponseWriterV2 tests", func() {
	Context("writing to response object", func() {
		response := NewProxyResponseWriterV2()

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
			resp := NewProxyResponseWriterV2()
			resp.Header().Add("Content-Type", "application/json")

			resp.Write([]byte(xmlBodyContent))

			Expect("application/json").To(Equal(resp.Header().Get("Content-Type")))
			proxyResp, err := resp.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(1).To(Equal(len(proxyResp.MultiValueHeaders)))
			Expect("application/json").To(Equal(proxyResp.MultiValueHeaders["Content-Type"][0]))
			Expect(xmlBodyContent).To(Equal(proxyResp.Body))
		})

		It("Sets the content type to text/xml given the body", func() {
			resp := NewProxyResponseWriterV2()
			resp.Write([]byte(xmlBodyContent))

			Expect("").ToNot(Equal(resp.Header().Get("Content-Type")))
			Expect(true).To(Equal(strings.HasPrefix(resp.Header().Get("Content-Type"), "text/xml;")))
			proxyResp, err := resp.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(1).To(Equal(len(proxyResp.MultiValueHeaders)))
			Expect(true).To(Equal(strings.HasPrefix(proxyResp.MultiValueHeaders["Content-Type"][0], "text/xml;")))
			Expect(xmlBodyContent).To(Equal(proxyResp.Body))
		})

		It("Sets the content type to text/html given the body", func() {
			resp := NewProxyResponseWriterV2()
			resp.Write([]byte(htmlBodyContent))

			Expect("").ToNot(Equal(resp.Header().Get("Content-Type")))
			Expect(true).To(Equal(strings.HasPrefix(resp.Header().Get("Content-Type"), "text/html;")))
			proxyResp, err := resp.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(1).To(Equal(len(proxyResp.MultiValueHeaders)))
			Expect(true).To(Equal(strings.HasPrefix(proxyResp.MultiValueHeaders["Content-Type"][0], "text/html;")))
			Expect(htmlBodyContent).To(Equal(proxyResp.Body))
		})
	})

	Context("Export API Gateway proxy response", func() {
		emtpyResponse := NewProxyResponseWriterV2()
		emtpyResponse.Header().Add("Content-Type", "application/json")

		It("Refuses empty responses with default status code", func() {
			_, err := emtpyResponse.GetProxyResponse()
			Expect(err).ToNot(BeNil())
			Expect("Status code not set on response").To(Equal(err.Error()))
		})

		simpleResponse := NewProxyResponseWriterV2()
		simpleResponse.Write([]byte("hello"))
		simpleResponse.Header().Add("Content-Type", "text/plain")
		It("Writes text body correctly", func() {
			proxyResponse, err := simpleResponse.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(proxyResponse).ToNot(BeNil())

			Expect("hello").To(Equal(proxyResponse.Body))
			Expect(http.StatusOK).To(Equal(proxyResponse.StatusCode))
			Expect(1).To(Equal(len(proxyResponse.MultiValueHeaders)))
			Expect(true).To(Equal(strings.HasPrefix(proxyResponse.MultiValueHeaders["Content-Type"][0], "text/plain")))
			Expect(proxyResponse.IsBase64Encoded).To(BeFalse())
		})

		binaryResponse := NewProxyResponseWriterV2()
		binaryResponse.Header().Add("Content-Type", "application/octet-stream")
		binaryBody := make([]byte, 256)
		_, err := rand.Read(binaryBody)
		if err != nil {
			Fail("Could not generate random binary body")
		}
		binaryResponse.Write(binaryBody)
		binaryResponse.WriteHeader(http.StatusAccepted)

		It("Encodes binary responses correctly", func() {
			proxyResponse, err := binaryResponse.GetProxyResponse()
			Expect(err).To(BeNil())
			Expect(proxyResponse).ToNot(BeNil())

			Expect(proxyResponse.IsBase64Encoded).To(BeTrue())
			Expect(base64.StdEncoding.EncodedLen(len(binaryBody))).To(Equal(len(proxyResponse.Body)))

			Expect(base64.StdEncoding.EncodeToString(binaryBody)).To(Equal(proxyResponse.Body))
			Expect(1).To(Equal(len(proxyResponse.MultiValueHeaders)))
			Expect("application/octet-stream").To(Equal(proxyResponse.MultiValueHeaders["Content-Type"][0]))
			Expect(http.StatusAccepted).To(Equal(proxyResponse.StatusCode))
		})
	})

	Context("Handle multi-value headers", func() {

		It("Writes single-value headers correctly", func() {
			response := NewProxyResponseWriterV2()
			response.Header().Add("Content-Type", "application/json")
			response.Write([]byte("hello"))
			proxyResponse, err := response.GetProxyResponse()
			Expect(err).To(BeNil())

			// Headers are not also written to `Headers` field
			Expect(0).To(Equal(len(proxyResponse.Headers)))
			Expect(1).To(Equal(len(proxyResponse.MultiValueHeaders["Content-Type"])))
			Expect("application/json").To(Equal(proxyResponse.MultiValueHeaders["Content-Type"][0]))
		})

		It("Writes multi-value headers correctly", func() {
			response := NewProxyResponseWriterV2()
			response.Header().Add("Set-Cookie", "csrftoken=foobar")
			response.Header().Add("Set-Cookie", "session_id=barfoo")
			response.Write([]byte("hello"))
			proxyResponse, err := response.GetProxyResponse()
			Expect(err).To(BeNil())

			// Headers are not also written to `Headers` field
			Expect(0).To(Equal(len(proxyResponse.Headers)))

			// There are two headers here because Content-Type is always written implicitly
			Expect(2).To(Equal(len(proxyResponse.MultiValueHeaders["Set-Cookie"])))
			Expect("csrftoken=foobar").To(Equal(proxyResponse.MultiValueHeaders["Set-Cookie"][0]))
			Expect("session_id=barfoo").To(Equal(proxyResponse.MultiValueHeaders["Set-Cookie"][1]))
		})
	})

})
