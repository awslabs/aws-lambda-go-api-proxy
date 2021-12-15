package handlerfunc

import (
	"net/http"

	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

type HandlerFuncAdapterV2 = httpadapter.HandlerAdapterV2

func NewV2(handlerFunc http.HandlerFunc) *HandlerFuncAdapterV2 {
	return httpadapter.NewV2(handlerFunc)
}
