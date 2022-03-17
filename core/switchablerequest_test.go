package core

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SwitchableAPIGatewayRequest", func() {
	Context("Serialization", func() {
		It("v1 serialized okay", func() {
			e := NewSwitchableAPIGatewayRequestV1(&events.APIGatewayProxyRequest{
				MultiValueQueryStringParameters: map[string][]string{},
			})
			b, err := json.Marshal(e)
			Expect(err).To(BeNil())
			m := map[string]interface{}{}
			err = json.Unmarshal(b, &m)
			Expect(err).To(BeNil())
			Expect(m["multiValueQueryStringParameters"]).To(Equal(map[string]interface {}{}))
			Expect(m["body"]).To(Equal(""))
		})
		It("v2 serialized okay", func() {
			e := NewSwitchableAPIGatewayRequestV2(&events.APIGatewayV2HTTPRequest{})
			b, err := json.Marshal(e)
			Expect(err).To(BeNil())
			m := map[string]interface{}{}
			err = json.Unmarshal(b, &m)
			Expect(err).To(BeNil())
			Expect(m["rawQueryString"]).To(Equal(""))
			Expect(m["isBase64Encoded"]).To(Equal(false))
		})
	})
	Context("Deserialization", func() {
		It("v1 deserialized okay", func() {
			input := &events.APIGatewayProxyRequest{
				Body:       "234",
				MultiValueQueryStringParameters: map[string][]string{
					"Test": []string{ "Value1", "Value2", },
				},
			}
			b, _ := json.Marshal(input)
			s := SwitchableAPIGatewayRequest{}
			err := s.UnmarshalJSON(b)
			Expect(err).To(BeNil())
			Expect(s.Version2()).To(BeNil())
			Expect(s.Version1()).To(BeEquivalentTo(input))
		})
		It("v2 deserialized okay", func() {
			input := &events.APIGatewayV2HTTPRequest{
				IsBase64Encoded:       true,
				RawQueryString: "a=b&c=d",
			}
			b, _ := json.Marshal(input)
			s := SwitchableAPIGatewayRequest{}
			err := s.UnmarshalJSON(b)
			Expect(err).To(BeNil())
			Expect(s.Version1()).To(BeNil())
			Expect(s.Version2()).To(BeEquivalentTo(input))
		})
	})})

func getProxyRequestV2(path string, method string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Path:   path,
				Method: method,
			},
		},
		RawPath: path,
	}
}
