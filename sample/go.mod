module github.com/awslabs/aws-lambda-go-api-proxy-sample

go 1.12

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/awslabs/aws-lambda-go-api-proxy v0.14.0
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.3.0
	github.com/kr/pretty v0.3.0 // indirect
	github.com/onsi/gomega v1.19.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.6.0
	gopkg.in/yaml.v2 v2.2.2 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.4 => gopkg.in/yaml.v2 v2.2.8
)
