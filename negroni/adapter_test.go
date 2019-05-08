package negroniadapter_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	negroniadapter "github.com/awslabs/aws-lambda-go-api-proxy/negroni"
	"github.com/urfave/negroni"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NegroniAdapter tests", func() {
	Context("Tests multiple handlers", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")

			homeHandler := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Home Page")
			}

			productsHandler := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("unfortunately-required-header", "")
				fmt.Fprintf(w, "Products Page")
			}

			mux := http.NewServeMux()
			mux.HandleFunc("/", homeHandler)
			mux.HandleFunc("/products", productsHandler)

			n := negroni.New()
			n.UseHandler(mux)

			adapter := negroniadapter.New(n)

			homePageReq := events.APIGatewayProxyRequest{
				Path:       "/",
				HTTPMethod: "GET",
			}

			homePageResp, homePageReqErr := adapter.ProxyWithContext(context.Background(), homePageReq)

			Expect(homePageReqErr).To(BeNil())
			Expect(homePageResp.StatusCode).To(Equal(200))
			Expect(homePageResp.Body).To(Equal("Home Page"))

			productsPageReq := events.APIGatewayProxyRequest{
				Path:       "/products",
				HTTPMethod: "GET",
			}

			productsPageResp, productsPageReqErr := adapter.Proxy(productsPageReq)

			Expect(productsPageReqErr).To(BeNil())
			Expect(productsPageResp.StatusCode).To(Equal(200))
			Expect(productsPageResp.Body).To(Equal("Products Page"))
		})
	})
})
