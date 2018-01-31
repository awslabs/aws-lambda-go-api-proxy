package negroniadapter

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/urfave/negroni"
)

type NegroniAdapter struct {
	core.RequestAccessor
	n *negroni.Negroni
}

func New(n *negroni.Negroni) *NegroniAdapter {
	return &NegroniAdapter{
		n: n,
	}
}

func newLoggedError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Println(err.Error())
	return err
}

func (h *NegroniAdapter) Proxy(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req, err := h.ProxyEventToHTTPRequest(event)
	if err != nil {
		return gatewayTimeout(), newLoggedError("Could not convert proxy event to request: %v", err)
	}

	w := core.NewProxyResponseWriter()
	h.n.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		return gatewayTimeout(), newLoggedError("Error while generating proxy response: %v", err)
	}

	return resp, nil
}

// Returns a dafault Gateway Timeout (504) response
func gatewayTimeout() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{StatusCode: http.StatusGatewayTimeout}
}
