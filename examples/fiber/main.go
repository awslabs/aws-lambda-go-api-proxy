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

var runOnAwsLambda = true
var fiberLambda *fiberadapter.FiberLambda
var app *fiber.App

// init the Fiber Server
func init() {
	log.Printf("Fiber cold start")

	app = fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	if runOnAwsLambda {
		log.Printf("Fiber Addapter New")
		fiberLambda = fiberadapter.New(app)
	}
}

// Handler will deal with Fiber working with Lambda
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return fiberLambda.ProxyWithContext(ctx, req)
}

func main() {

	if runOnAwsLambda {
		// Make the handler available for Remote Procedure Call by AWS Lambda
		log.Printf("Fiber Adapter Lambda Handler")
		lambda.Start(Handler)
	} else {
		log.Printf("Fiber Listen Handler")
		log.Fatal(app.Listen(":3000"))
	}
}
