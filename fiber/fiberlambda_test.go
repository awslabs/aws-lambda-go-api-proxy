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
			log.Println("Starting test")

			app := fiber.New()
			app.Get("/ping", func(c *fiber.Ctx) error {
				c.Next()
				Expect(c.Get(fiber.HeaderUserAgent)).To(Equal("fiber"))
				Expect(c.Get(fiber.HeaderContentType)).To(Equal(fiber.MIMEApplicationJSONCharsetUTF8))
				Expect(c.Get(fiber.HeaderReferer)).To(Equal("https://github.com/gofiber/fiber"))
				Expect(c.Get(fiber.HeaderConnection)).To(Equal("Keep-Alive"))
				Expect(c.Get(fiber.HeaderKeepAlive)).To(Equal("timeout=5, max=1000"))

				Expect(c.Cookies("a")).To(Equal("b"))
				Expect(c.Cookies("b")).To(Equal("c"))
				Expect(c.Cookies("c")).To(Equal("d"))

				c.Context().Request.Header.VisitAll(func(key, value []byte) {
					if string(key) == "K1" {
						Expect(c.Get("K1")).To(Or(Equal("v1"), Equal("v2")))
					}
				})
				return c.SendString("pong")
			})

			adapter := fiberadaptor.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/ping",
				HTTPMethod: "GET",
				MultiValueHeaders: map[string][]string{
					fiber.HeaderConnection:       {"Keep-Alive"},
					fiber.HeaderKeepAlive:        {"timeout=5, max=1000"},
					fiber.HeaderUserAgent:        {"fiber"},
					fiber.HeaderContentType:      {fiber.MIMEApplicationJSONCharsetUTF8},
					fiber.HeaderReferer:          {"https://github.com/gofiber/fiber"},
					"k1":                         {"v1", "v2"},
					"cookie":                     {"a=b", "b=c;c=d"},
				},
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Body).To(Equal("pong"))
		})
	})
})
