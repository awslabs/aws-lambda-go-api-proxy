package ginadapter_test

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GinLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.New(r)

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
	Context("Function URL", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.New(r)

			req := events.LambdaFunctionURLRequest{
				RawPath: "/ping",
			}

			resp, err := adapter.ProxyFunctionURL(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Body).To(Equal("{\"message\":\"pong\"}"))
		})

		It("Proxies the event correctly with context", func() {
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.New(r)

			req := events.LambdaFunctionURLRequest{
				RawPath: "/ping",
			}

			ctx := context.Background()
			resp, err := adapter.ProxyFunctionURLWithContext(ctx, req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Body).To(Equal("{\"message\":\"pong\"}"))
		})
	})
})

var _ = Describe("GinLambdaV2 tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.NewV2(r)

			req := events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: "GET",
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

var _ = Describe("GinLambdaALB tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.NewALB(r)

			req := events.ALBTargetGroupRequest{
				HTTPMethod: "GET",
				Path:       "/ping",
				RequestContext: events.ALBTargetGroupRequestContext{
					ELB: events.ELBContext{TargetGroupArn: " ad"},
				}}

			resp, err := adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))

			resp, err = adapter.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})
