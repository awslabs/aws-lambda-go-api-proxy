package core

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
)

// SwitchableAPIGatewayResponse is a container for an APIGatewayProxyResponse or an APIGatewayV2HTTPResponse object which
// handles serialization and deserialization and switching between the entities based on the presence of fields in the
// source JSON, multiValueQueryStringParameters for APIGatewayProxyResponse and rawQueryString for
// APIGatewayV2HTTPResponse. It also provides some simple switching functions (wrapped type switching.)
type SwitchableAPIGatewayResponse struct {
	v interface{}
}

// NewSwitchableAPIGatewayResponseV1 creates a new SwitchableAPIGatewayResponse from APIGatewayProxyResponse
func NewSwitchableAPIGatewayResponseV1(v *events.APIGatewayProxyResponse) *SwitchableAPIGatewayResponse {
	return &SwitchableAPIGatewayResponse{
		v: v,
	}
}

// NewSwitchableAPIGatewayResponseV2 creates a new SwitchableAPIGatewayResponse from APIGatewayV2HTTPResponse
func NewSwitchableAPIGatewayResponseV2(v *events.APIGatewayV2HTTPResponse) *SwitchableAPIGatewayResponse {
	return &SwitchableAPIGatewayResponse{
		v: v,
	}
}

// MarshalJSON is a pass through serialization
func (s *SwitchableAPIGatewayResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

// UnmarshalJSON is a switching serialization based on the presence of fields in the
// source JSON, statusCode to verify that it's either APIGatewayProxyResponse or APIGatewayV2HTTPResponse and then
// rawQueryString for to determine if it is APIGatewayV2HTTPResponse or not.
func (s *SwitchableAPIGatewayResponse) UnmarshalJSON(b []byte) error {
	delta := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &delta); err != nil {
		return err
	}
	_, test := delta["statusCode"]
	_, v2test := delta["cookies"]
	s.v = nil
	if test && !v2test {
		s.v = &events.APIGatewayProxyResponse{}
	} else if test && v2test {
		s.v = &events.APIGatewayV2HTTPResponse{}
	} else {
		return errors.New("unable to determine response version")
	}
	return json.Unmarshal(b, s.v)
}

// Version1 returns the contained events.APIGatewayProxyResponse or nil
func (s *SwitchableAPIGatewayResponse) Version1() *events.APIGatewayProxyResponse {
	switch v := s.v.(type) {
	case *events.APIGatewayProxyResponse:
		return v
	case events.APIGatewayProxyResponse:
		return &v
	}
	return nil
}

// Version2 returns the contained events.APIGatewayV2HTTPResponse or nil
func (s *SwitchableAPIGatewayResponse) Version2() *events.APIGatewayV2HTTPResponse {
	switch v := s.v.(type) {
	case *events.APIGatewayV2HTTPResponse:
		return v
	case events.APIGatewayV2HTTPResponse:
		return &v
	}
	return nil
}
