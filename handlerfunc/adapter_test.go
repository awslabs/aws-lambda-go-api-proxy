package handlerfunc_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/handlerfunc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("unfortunately-required-header", "")
	fmt.Fprintf(w, "Go Lambda!!")
}

var _ = Describe("HandlerFuncAdapter tests", func() {
	Context("Simple ping request, HandlerFunc", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")

			handlerFunc := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Go Lambda!!")
			}

			adapter := handlerfunc.New(handlerFunc)

			req := events.APIGatewayProxyRequest{
				Path:       "/ping",
				HTTPMethod: "GET",
			}

			resp, err := adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})

	Context("Simple ping request, Handler", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")

			adapter := handlerfunc.NewHandler(handler{})

			req := events.APIGatewayProxyRequest{
				Path:       "/ping",
				HTTPMethod: "GET",
			}

			resp, err := adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})
