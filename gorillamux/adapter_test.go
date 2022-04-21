package gorillamux_test

import (
	"context"
	"fmt"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GorillaMuxAdapter tests", func() {
	Context("Simple ping request with v1", func() {
		It("Proxies the event correctly", func() {
			homeHandler := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Home Page")
			}

			productsHandler := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Products Page")
			}

			r := mux.NewRouter()
			r.HandleFunc("/", homeHandler)
			r.HandleFunc("/products", productsHandler)

			adapter := gorillamux.New(r)

			homePageReq := events.APIGatewayProxyRequest{
				Path:       "/",
				HTTPMethod: "GET",
			}

			homePageResp, homePageReqErr := adapter.ProxyWithContext(context.Background(), *core.NewSwitchableAPIGatewayRequestV1(&homePageReq))

			Expect(homePageReqErr).To(BeNil())
			Expect(homePageResp.Version1().StatusCode).To(Equal(200))
			Expect(homePageResp.Version1().Body).To(Equal("Home Page"))

			productsPageReq := events.APIGatewayProxyRequest{
				Path:       "/products",
				HTTPMethod: "GET",
			}

			productsPageResp, productsPageReqErr := adapter.Proxy(*core.NewSwitchableAPIGatewayRequestV1(&productsPageReq))

			Expect(productsPageReqErr).To(BeNil())
			Expect(productsPageResp.Version1().StatusCode).To(Equal(200))
			Expect(productsPageResp.Version1().Body).To(Equal("Products Page"))
		})
	})

	Context("Simple ping request with v2", func() {
		It("Proxies the event correctly", func() {
			homeHandler := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Home Page")
			}

			productsHandler := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Products Page")
			}

			r := mux.NewRouter()
			r.HandleFunc("/", homeHandler)
			r.HandleFunc("/products", productsHandler)

			adapter := gorillamux.New(r)

			homePageReq := getProxyRequestV2("/", "GET")

			homePageResp, homePageReqErr := adapter.ProxyWithContext(context.Background(), *core.NewSwitchableAPIGatewayRequestV2(&homePageReq))

			Expect(homePageReqErr).To(BeNil())
			Expect(homePageResp.Version2().StatusCode).To(Equal(200))
			Expect(homePageResp.Version2().Body).To(Equal("Home Page"))

			productsPageReq := getProxyRequestV2("/products", "GET")

			productsPageResp, productsPageReqErr := adapter.Proxy(*core.NewSwitchableAPIGatewayRequestV2(&productsPageReq))

			Expect(productsPageReqErr).To(BeNil())
			Expect(productsPageResp.Version2().StatusCode).To(Equal(200))
			Expect(productsPageResp.Version2().Body).To(Equal("Products Page"))
		})
	})
})

func getProxyRequestV2(path string, method string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Path:   path,
				Method: method,
			},
		},
		RawPath: path,
	}
}

