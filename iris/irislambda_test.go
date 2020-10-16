package irisadapter_test

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	irisadapter "github.com/awslabs/aws-lambda-go-api-proxy/iris"
	"github.com/kataras/iris/v12"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IrisLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")

			app := iris.New()
			app.Get("/ping", func(ctx iris.Context) {
				log.Println("Handler!!")
				ctx.WriteString("pong")
			})

			adapter := irisadapter.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/ping",
				HTTPMethod: "GET",
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
