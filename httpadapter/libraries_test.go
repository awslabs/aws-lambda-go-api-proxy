package httpadapter_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	"github.com/kataras/iris/v12"
	"github.com/labstack/echo/v4"
	"github.com/urfave/negroni"
	"goji.io"
	"goji.io/pat"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPAdapter tests", func() {
	Context("Simple ping request", func() {
		tests := []struct {
			name    string
			handler http.Handler
		}{
			{
				name: "chi",
				handler: func() http.Handler {
					r := chi.NewRouter()
					r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprintf(w, "pong")
					})

					return r
				}(),
			},
			{
				name: "echo",
				handler: func() http.Handler {
					e := echo.New()
					e.GET("/ping", func(c echo.Context) error {
						return c.String(200, "pong")
					})

					return e
				}(),
			},
			{
				name: "gin",
				handler: func() http.Handler {
					gin.SetMode(gin.ReleaseMode)

					r := gin.New()
					r.GET("/ping", func(c *gin.Context) {
						c.String(200, "pong")
					})

					return r
				}(),
			},
			{
				name: "goji",
				handler: func() http.Handler {
					r := goji.NewMux()
					r.HandleFunc(pat.Get("/ping"), func(w http.ResponseWriter, req *http.Request) {
						fmt.Fprintf(w, "pong")
					})

					return r
				}(),
			},
			{
				name: "gorillamux",
				handler: func() http.Handler {
					r := mux.NewRouter()
					r.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
						fmt.Fprintf(w, "pong")
					})

					return r
				}(),
			},
			{
				name: "iris",
				handler: func() http.Handler {
					app := iris.Default()
					app.Get("/ping", func(ctx iris.Context) {
						log.Println("Handler!!")
						ctx.WriteString("pong")
					})
					app.Build()

					return app
				}(),
			},
			{
				name: "negroni",
				handler: func() http.Handler {
					mux := http.NewServeMux()
					mux.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
						fmt.Fprintf(w, "pong")
					})

					n := negroni.New()
					n.UseHandler(mux)

					return n
				}(),
			},
		}

		req := events.APIGatewayProxyRequest{
			Path:       "/ping",
			HTTPMethod: "GET",
		}

		for i := range tests {
			test := tests[i]
			It("Proxies the event to "+test.name, func() {
				adapter := httpadapter.New(test.handler)

				resp, err := adapter.ProxyWithContext(context.Background(), req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(200))
				Expect(resp.Body).To(Equal("pong"))

				resp, err = adapter.Proxy(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(200))
				Expect(resp.Body).To(Equal("pong"))
			})
		}
	})
})
