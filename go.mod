module github.com/ylck/aws-lambda-go-api-proxy

go 1.14

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/aws/aws-lambda-go v1.19.1
	github.com/aymerick/raymond v2.0.3-0.20180322193309-b565731e1464+incompatible // indirect
	github.com/gin-gonic/gin v1.7.7
	github.com/go-chi/chi/v5 v5.0.2
	github.com/goccy/go-json v0.9.7 // indirect
	github.com/gofiber/fiber/v2 v2.1.0
	github.com/gorilla/mux v1.7.4
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/kataras/iris/v12 v12.2.0-alpha9
	github.com/kataras/tunnel v0.0.4 // indirect
	github.com/klauspost/compress v1.15.6 // indirect
	github.com/labstack/echo/v4 v4.9.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/tdewolff/minify/v2 v2.11.10 // indirect
	github.com/urfave/negroni v1.0.0
	github.com/valyala/fasthttp v1.34.0
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gopkg.in/yaml.v2 v2.2.2 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.3 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.4 => gopkg.in/yaml.v2 v2.2.8
)
