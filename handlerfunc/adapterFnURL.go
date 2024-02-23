package handlerfunc

import (
	"net/http"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

type HandlerFuncAdapterFnURL = httpadapter.HandlerAdapterFnURL

func NewFunctionURL(handlerFunc http.HandlerFunc) *HandlerFuncAdapterFnURL {
	return httpadapter.NewFunctionURL(handlerFunc)
}
