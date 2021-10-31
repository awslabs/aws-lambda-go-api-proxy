package core

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
)

type SwitchableAPIGatewayRequest struct {
	v interface{} // v is Always nil, or a pointer of APIGatewayProxyRequest or APIGatewayV2HTTPRequest
}

// NewSwitchableAPIGatewayRequestV1 creates a new SwitchableAPIGatewayRequest from APIGatewayProxyRequest
func NewSwitchableAPIGatewayRequestV1(v *events.APIGatewayProxyRequest) *SwitchableAPIGatewayRequest {
	return &SwitchableAPIGatewayRequest{
		v: v,
	}
}
// NewSwitchableAPIGatewayRequestV2 creates a new SwitchableAPIGatewayRequest from APIGatewayV2HTTPRequest
func NewSwitchableAPIGatewayRequestV2(v *events.APIGatewayV2HTTPRequest) *SwitchableAPIGatewayRequest {
	return &SwitchableAPIGatewayRequest{
		v: v,
	}
}

// MarshalJSON is a pass through serialization
func (s *SwitchableAPIGatewayRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

// UnmarshalJSON is a switching serialization based on the presence of fields in the
// source JSON, multiValueQueryStringParameters for APIGatewayProxyRequest and rawQueryString for
// APIGatewayV2HTTPRequest.
func (s *SwitchableAPIGatewayRequest) UnmarshalJSON(b []byte) error {
	delta := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &delta); err != nil {
		return err
	}
	_, v1test := delta["multiValueQueryStringParameters"]
	_, v2test := delta["rawQueryString"]
	s.v = nil
	if v1test && !v2test {
		s.v = &events.APIGatewayProxyRequest{}
	} else if !v1test && v2test {
		s.v = &events.APIGatewayV2HTTPRequest{}
	} else {
		return errors.New("unable to determine request version")
	}
	return json.Unmarshal(b, s.v)
}

// Version1 returns the contained events.APIGatewayProxyRequest or nil
func (s *SwitchableAPIGatewayRequest) Version1() *events.APIGatewayProxyRequest {
	switch v := s.v.(type) {
	case *events.APIGatewayProxyRequest:
		return v
	case events.APIGatewayProxyRequest:
		return &v
	}
	return nil
}

// Version2 returns the contained events.APIGatewayV2HTTPRequest or nil
func (s *SwitchableAPIGatewayRequest) Version2() *events.APIGatewayV2HTTPRequest {
	switch v := s.v.(type) {
	case *events.APIGatewayV2HTTPRequest:
		return v
	case events.APIGatewayV2HTTPRequest:
		return &v
	}
	return nil
}
