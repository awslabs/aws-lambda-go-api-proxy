package ginlambda_test

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
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

			lambdaGin := ginlambda.New(r)

			req := events.APIGatewayProxyRequest{
				Path:       "/ping",
				HTTPMethod: "GET",
			}

			resp, err := lambdaGin.Proxy(req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})
