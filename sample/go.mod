module github.com/awslabs/aws-lambda-go-api-proxy-sample

go 1.12

require (
	github.com/aws/aws-lambda-go v1.10.0
	github.com/awslabs/aws-lambda-go-api-proxy v0.3.0
	github.com/gin-gonic/gin v1.7.0
	github.com/google/uuid v1.1.1
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.19.0 // indirect
)

replace (
	gopkg.in/yaml.v2 v2.2.2 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.4 => gopkg.in/yaml.v2 v2.2.8
)
