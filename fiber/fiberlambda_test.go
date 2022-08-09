package fiberadapter_test

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gofiber/fiber/v2"

	fiberadaptor "github.com/awslabs/aws-lambda-go-api-proxy/fiber"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FiberLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			app := fiber.New()
			app.Get("/ping", func(c *fiber.Ctx) error {
				return c.SendString("pong")
			})

			adapter := fiberadaptor.New(app)

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

	Context("Request header", func() {
		It("Check pass canonical header to fiber", func() {
			app := fiber.New()
			app.Post("/canonical_header", func(c *fiber.Ctx) error {
				Expect(c.Get(fiber.HeaderHost)).To(Equal("localhost"))
				Expect(c.Get(fiber.HeaderContentType)).To(Equal(fiber.MIMEApplicationJSONCharsetUTF8))
				Expect(c.Get(fiber.HeaderUserAgent)).To(Equal("fiber"))

				Expect(c.Cookies("a")).To(Equal("b"))
				Expect(c.Cookies("b")).To(Equal("c"))
				Expect(c.Cookies("c")).To(Equal("d"))

				Expect(c.Get(fiber.HeaderContentLength)).To(Equal("77"))
				Expect(c.Get(fiber.HeaderConnection)).To(Equal("Keep-Alive"))
				Expect(c.Get(fiber.HeaderKeepAlive)).To(Equal("timeout=5, max=1000"))

				return c.Status(fiber.StatusNoContent).Send(nil)
			})

			adapter := fiberadaptor.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/canonical_header",
				HTTPMethod: "POST",
				MultiValueHeaders: map[string][]string{
					fiber.HeaderHost:        {"localhost"},
					fiber.HeaderContentType: {fiber.MIMEApplicationJSONCharsetUTF8},
					fiber.HeaderUserAgent:   {"fiber"},

					"cookie": {"a=b", "b=c;c=d"},

					fiber.HeaderContentLength: {"77"},
					fiber.HeaderConnection:    {"Keep-Alive"},
					fiber.HeaderKeepAlive:     {"timeout=5, max=1000"},
				},
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(fiber.StatusNoContent))
			Expect(resp.Body).To(Equal(""))
		})

		It("Check pass non canonical header to fiber", func() {
			app := fiber.New()
			app.Post("/header", func(c *fiber.Ctx) error {
				Expect(c.Get(fiber.HeaderReferer)).To(Equal("https://github.com/gofiber/fiber"))
				Expect(c.Get(fiber.HeaderAuthorization)).To(Equal("Bearer drink_beer_not_coffee"))

				c.Context().Request.Header.VisitAll(func(key, value []byte) {
					if string(key) == "K1" {
						Expect(Expect(c.Get("K1")).To(Or(Equal("v1"), Equal("v2"))))
					}
				})

				return c.Status(fiber.StatusNoContent).Send(nil)
			})

			adapter := fiberadaptor.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/header",
				HTTPMethod: "POST",
				MultiValueHeaders: map[string][]string{
					fiber.HeaderReferer:       {"https://github.com/gofiber/fiber"},
					fiber.HeaderAuthorization: {"Bearer drink_beer_not_coffee"},

					"k1": {"v1", "v2"},
				},
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(fiber.StatusNoContent))
			Expect(resp.Body).To(Equal(""))
		})
	})

	Context("Response header", func() {
		It("Check pass canonical header to fiber", func() {
			app := fiber.New()
			app.Post("/canonical_header", func(c *fiber.Ctx) error {
				c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
				c.Set(fiber.HeaderServer, "localhost")

				c.Cookie(&fiber.Cookie{
					Name:     "a",
					Value:    "b",
					HTTPOnly: true,
				})
				c.Cookie(&fiber.Cookie{
					Name:     "b",
					Value:    "c",
					HTTPOnly: true,
				})
				c.Cookie(&fiber.Cookie{
					Name:     "c",
					Value:    "d",
					HTTPOnly: true,
				})

				c.Set(fiber.HeaderContentLength, "77")
				c.Set(fiber.HeaderConnection, "keep-alive")

				return c.Status(fiber.StatusNoContent).Send(nil)
			})

			adapter := fiberadaptor.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/canonical_header",
				HTTPMethod: "POST",
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(fiber.StatusNoContent))
			// NOTI: core.NewProxyResponseWriter().GetProxyResponse() => Doesn't use `resp.Header`
			Expect(resp.MultiValueHeaders[fiber.HeaderContentType]).To(Equal([]string{fiber.MIMEApplicationJSONCharsetUTF8}))
			Expect(resp.MultiValueHeaders[fiber.HeaderServer]).To(Equal([]string{"localhost"}))
			Expect(resp.MultiValueHeaders[fiber.HeaderSetCookie]).To(Equal([]string{"a=b; path=/; HttpOnly; SameSite=Lax", "b=c; path=/; HttpOnly; SameSite=Lax", "c=d; path=/; HttpOnly; SameSite=Lax"}))
			Expect(resp.MultiValueHeaders[fiber.HeaderContentLength]).To(Equal([]string{"77"}))
			Expect(resp.MultiValueHeaders[fiber.HeaderConnection]).To(Equal([]string{"keep-alive"}))
			Expect(resp.Body).To(Equal(""))
		})
		It("Check pass non canonical header to fiber", func() {
			app := fiber.New()
			app.Post("/header", func(c *fiber.Ctx) error {
				c.Links("http://api.example.com/users?page=2", "next", "http://api.example.com/users?page=5", "last")
				return c.Redirect("https://github.com/gofiber/fiber")
			})

			adapter := fiberadaptor.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/header",
				HTTPMethod: "POST",
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(fiber.StatusFound))
			Expect(resp.MultiValueHeaders[fiber.HeaderLocation]).To(Equal([]string{"https://github.com/gofiber/fiber"}))
			Expect(resp.MultiValueHeaders[fiber.HeaderLink]).To(Equal([]string{"<http://api.example.com/users?page=2>; rel=\"next\",<http://api.example.com/users?page=5>; rel=\"last\""}))
			Expect(resp.Body).To(Equal(""))
		})
	})

	Context("Next method", func() {
		It("Check missing values in request header", func() {
			app := fiber.New()
			app.Post("/next", func(c *fiber.Ctx) error {
				c.Next()
				Expect(c.Get(fiber.HeaderHost)).To(Equal("localhost"))
				Expect(c.Get(fiber.HeaderContentType)).To(Equal(fiber.MIMEApplicationJSONCharsetUTF8))
				Expect(c.Get(fiber.HeaderUserAgent)).To(Equal("fiber"))

				Expect(c.Cookies("a")).To(Equal("b"))
				Expect(c.Cookies("b")).To(Equal("c"))
				Expect(c.Cookies("c")).To(Equal("d"))

				Expect(c.Get(fiber.HeaderContentLength)).To(Equal("77"))
				Expect(c.Get(fiber.HeaderConnection)).To(Equal("Keep-Alive"))
				Expect(c.Get(fiber.HeaderKeepAlive)).To(Equal("timeout=5, max=1000"))

				return c.Status(fiber.StatusNoContent).Send(nil)
			})
			adapter := fiberadaptor.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/next",
				HTTPMethod: "POST",
				MultiValueHeaders: map[string][]string{
					fiber.HeaderHost:        {"localhost"},
					fiber.HeaderContentType: {fiber.MIMEApplicationJSONCharsetUTF8},
					fiber.HeaderUserAgent:   {"fiber"},

					"cookie": {"a=b", "b=c;c=d"},

					fiber.HeaderContentLength: {"77"},
					fiber.HeaderConnection:    {"Keep-Alive"},
					fiber.HeaderKeepAlive:     {"timeout=5, max=1000"},
				},
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(fiber.StatusNoContent))
			Expect(resp.Body).To(Equal(""))
		})

		It("Check missing values in response header", func() {
			app := fiber.New()
			app.Post("/next", func(c *fiber.Ctx) error {
				c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
				c.Set(fiber.HeaderServer, "localhost")

				c.Cookie(&fiber.Cookie{
					Name:     "a",
					Value:    "b",
					HTTPOnly: true,
				})
				c.Cookie(&fiber.Cookie{
					Name:     "b",
					Value:    "c",
					HTTPOnly: true,
				})
				c.Cookie(&fiber.Cookie{
					Name:     "c",
					Value:    "d",
					HTTPOnly: true,
				})

				c.Set(fiber.HeaderContentLength, "77")
				c.Set(fiber.HeaderConnection, "keep-alive")

				c.Next()
				return c.Status(fiber.StatusNoContent).Send(nil)
			})
			adapter := fiberadaptor.New(app)

			req := events.APIGatewayProxyRequest{
				Path:       "/next",
				HTTPMethod: "POST",
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(fiber.StatusNoContent))
			Expect(resp.MultiValueHeaders[fiber.HeaderContentType]).To(Equal([]string{fiber.MIMEApplicationJSONCharsetUTF8}))
			Expect(resp.MultiValueHeaders[fiber.HeaderServer]).To(Equal([]string{"localhost"}))
			Expect(resp.MultiValueHeaders[fiber.HeaderSetCookie]).To(Equal([]string{"a=b; path=/; HttpOnly; SameSite=Lax", "b=c; path=/; HttpOnly; SameSite=Lax", "c=d; path=/; HttpOnly; SameSite=Lax"}))
			Expect(resp.MultiValueHeaders[fiber.HeaderContentLength]).To(Equal([]string{"77"}))
			Expect(resp.MultiValueHeaders[fiber.HeaderConnection]).To(Equal([]string{"keep-alive"}))
			Expect(resp.Body).To(Equal(""))
		})
	})
})
