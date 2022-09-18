package echoadapter_test

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EchoLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			e := echo.New()
			e.GET("/ping", func(c echo.Context) error {
				log.Println("Handler!!")
				return c.String(200, "pong")
			})

			adapter := echoadapter.New(e)

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

var _ = Describe("EchoLambdaV2 tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			e := echo.New()
			e.GET("/ping", func(c echo.Context) error {
				log.Println("Handler!!")
				return c.String(200, "pong")
			})

			adapter := echoadapter.NewV2(e)

			req := events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: "GET",
						Path:   "/ping",
					},
				},
			}

			resp, err := adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})

var _ = Describe("EchoLambdaALB tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			e := echo.New()
			e.GET("/ping", func(c echo.Context) error {
				log.Println("Handler!!")
				return c.String(200, "pong")
			})

			adapter := echoadapter.NewALB(e)

			req := events.ALBTargetGroupRequest{
				HTTPMethod:     "GET",
				Path:           "/ping",
				RequestContext: events.ALBTargetGroupRequestContext{},
			}

			resp, err := adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})
