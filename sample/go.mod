module github.com/awslabs/aws-lambda-go-api-proxy-sample

go 1.12

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/awslabs/aws-lambda-go-api-proxy v0.14.0
	github.com/bytedance/sonic v1.8.2 // indirect
	github.com/gin-gonic/gin v1.9.0
	github.com/google/uuid v1.3.0
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.2 // indirect
	github.com/onsi/gomega v1.19.0 // indirect
	github.com/ugorji/go/codec v1.2.10 // indirect
	golang.org/x/arch v0.2.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.6.0
	gopkg.in/yaml.v2 v2.2.2 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v2 v2.2.4 => gopkg.in/yaml.v2 v2.2.8
)
