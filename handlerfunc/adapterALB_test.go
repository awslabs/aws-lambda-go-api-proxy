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

var _ = Describe("HandlerFuncAdapter ALB tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")

			handler := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Go Lambda!!")
			}

			adapter := handlerfunc.NewALB(handler)

			req := events.ALBTargetGroupRequest{
				HTTPMethod: http.MethodGet,
				Path:       "/",
				RequestContext: events.ALBTargetGroupRequestContext{
					ELB: events.ELBContext{TargetGroupArn: " ad"},
				}}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))

			resp, err = adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})
