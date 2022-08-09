## AWS Lambda Go API Proxy [![Build Status](https://travis-ci.org/awslabs/aws-lambda-go-api-proxy.svg?branch=master)](https://travis-ci.org/awslabs/aws-lambda-go-api-proxy)
aws-lambda-go-api-proxy makes it easy to run Go APIs written with frameworks such as [Gin](https://github.com/gin-gonic/gin) with AWS Lambda and Amazon API Gateway.

## Getting started

Install required dependencies.

```bash
# First, install the Lambda go libraries.
$ go get github.com/aws/aws-lambda-go/events
$ go get github.com/aws/aws-lambda-go/lambda

# Next, install the core library.
$ go get github.com/awslabs/aws-lambda-go-api-proxy/...
```

### Standard library

To use with the standard library, the `httpadaptor.New` function takes in a `http.Handler`. The `ProxyWithContent` method on the `httpadapter.HandlerAdapter` can then be used as a Lambda handler.

```go
package main

import (
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	})

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
}
```

### Gin

To use with the Gin framework, following the instructions from the [Lambda documentation](https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html), declare a `Handler` method for the main package.

Declare a `ginadapter.GinLambda` object in the global scope, and initialize it in the `init` function, adding all API methods.

The `ProxyWithContext` method is then used to translate requests and responses.

```go
package main

import (
	"log"
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Gin cold start")
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
```

### Fiber

To use with the Fiber framework, following the instructions from the [Lambda documentation](https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html), declare a `Handler` method for the main package.

Declare a `fiberadapter.FiberLambda` object in the global scope, and initialize it in the `init` function, adding all API methods.

The `ProxyWithContext` method is then used to translate requests and responses.

```go
// main.go
package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	fiberadapter "github.com/awslabs/aws-lambda-go-api-proxy/fiber"
	"github.com/gofiber/fiber/v2"
)

var fiberLambda *fiberadapter.FiberLambda

// init the Fiber Server
func init() {
	log.Printf("Fiber cold start")
	var app *fiber.App
	app = fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	fiberLambda = fiberadapter.New(app)
}

// Handler will deal with Fiber working with Lambda
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return fiberLambda.ProxyWithContext(ctx, req)
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(Handler)
}
```

## Other frameworks
This package also supports [Negroni](https://github.com/urfave/negroni), [GorillaMux](https://github.com/gorilla/mux), and plain old `HandlerFunc` - take a look at the code in their respective sub-directories. All packages implement the `Proxy` method exactly like our Gin sample above.

## Deploying the sample
We have included a [SAM template](https://github.com/awslabs/serverless-application-model) with our sample application. You can use the [AWS CLI](https://aws.amazon.com/cli/) to quickly deploy the application in your AWS account.

First, build the sample application by running `make` from the `aws-lambda-go-api-proxy` directory.

```bash
$ cd aws-lambda-go-api-proxy
$ make
```

The `make` process should generate a `main.zip` file in the sample folder. You can now use the AWS CLI to prepare the deployment for AWS Lambda and Amazon API Gateway.

```bash
$ cd sample
$ aws cloudformation package --template-file sam.yaml --output-template-file output-sam.yaml --s3-bucket YOUR_DEPLOYMENT_BUCKET
$ aws cloudformation deploy --template-file output-sam.yaml --stack-name YOUR_STACK_NAME --capabilities CAPABILITY_IAM
```

Using the CloudFormation console, you can find the URL for the newly created API endpoint in the `Outputs` tab of the sample stack - it looks sample like this: `https://xxxxxxxxx.execute-api.xx-xxxx-x.amazonaws.com/Prod/pets`. Open a browser window and try to call the URL.

## API Gateway context and stage variables
~~The `RequestAccessor` object, and therefore `GinLambda`, automatically marshals the API Gateway request context and stage variables objects and stores them in custom headers in the request: `X-GinLambda-ApiGw-Context` and `X-GinLambda-ApiGw-StageVars`. While you could manually unmarshal the json content into the `events.APIGatewayProxyRequestContext` and `map[string]string` objects, the library exports two utility methods to give you easy access to the data.~~

The gateway context, stage variables and lambda runtime variables are automatically populate to the context.

```go
// the methods are available in your instance of the GinLambda
// object and receive the context
apiGwContext := ginLambda.GetAPIGatewayContextFromContext(ctx)
apiGwStageVars := ginLambda.GetStageVarsFromContext(ctx)
runtimeContext := ginLambda.GetRuntimeContextFromContext(ctx)

// you can access the properties of the context directly
log.Println(apiGwContext.RequestID)
log.Println(apiGwContext.Stage)
log.Println(runtimeContext.InvokedFunctionArn)


// stage variables are stored in a map[string]string
stageVarValue := apiGwStageVars["MyStageVar"]
```

## Supporting other frameworks
The `aws-lambda-go-api-proxy`, alongside the various adapters, declares a `core` package. The `core` package, contains utility methods and interfaces to translate API Gateway proxy events into Go's default `http.Request` and `http.ResponseWriter` objects.

You can see that the [`ginlambda.go`](gin/adapter.go) file extends the `RequestAccessor` struct defined in the [`request.go`](core/request.go) file.  `RequestAccessor` gives you access to the `ProxyEventToHTTPRequest()` method.

The `GinLambda` object is initialized with an instance of `gin.Engine`. `gin.Engine` implements methods defined in the `http.Handler` interface.

The `Proxy` method of the `GinLambda` object simply receives the `events.APIGatewayProxyRequest` object and uses the `ProxyEventToHTTPRequest()` method to convert it into an `http.Request` object. Next, it creates a new `ProxyResponseWriter` object (defined in the [`response.go`](core/response.go)) file and passes both request and response writer to the `ServeHTTP` method of the `gin.Engine`.

The `ProxyResponseWriter` exports a method called `GetProxyResponse()` to generate an `events.APIGatewayProxyResponse` object from the data written to the response writer.

Support for frameworks other than Gin can rely on the same methods from the `core` package and swap the `gin.Engine` object for the relevant framework's object.

## License

This library is licensed under the Apache 2.0 License.
