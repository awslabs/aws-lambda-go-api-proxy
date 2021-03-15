package handlerfunc_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/handlerfunc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandlerFuncAdapter tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")

			handler := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Go Lambda!!")
			}

			adapter := handlerfunc.NewV2(handler)

			req := events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: http.MethodGet,
						Path:   "/ping",
					},
				},
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))

			resp, err = adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})
