package fiberadapter_test

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	fiberadaptor "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/gofiber/fiber/v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FiberLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			const AgentGolang = "Agent Gopher/1.16"
			log.Println("Starting test")
			app := fiber.New()
			app.Get("/ping", func(c *fiber.Ctx) error {
				log.Println("Handler!!")
				Expect(c.Get(fiber.HeaderUserAgent, "")).To(Equal(AgentGolang))
				Expect(c.Get(fiber.HeaderContentType, "")).To(Equal(fiber.MIMEApplicationJSONCharsetUTF8))
				return c.SendString("pong")
			})

			adapter := fiberadaptor.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/ping",
				HTTPMethod: "GET",
				Headers: map[string]string{
					fiber.HeaderUserAgent:   AgentGolang,
					fiber.HeaderContentType: fiber.MIMEApplicationJSONCharsetUTF8,
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
