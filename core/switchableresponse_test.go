package core

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SwitchableAPIGatewayResponse", func() {
	Context("Serialization", func() {
		It("v1 serialized okay", func() {
			e := NewSwitchableAPIGatewayResponseV1(&events.APIGatewayProxyResponse{})
			b, err := json.Marshal(e)
			Expect(err).To(BeNil())
			m := map[string]interface{}{}
			err = json.Unmarshal(b, &m)
			Expect(err).To(BeNil())
			Expect(m["statusCode"]).To(Equal(0.0))
			Expect(m["body"]).To(Equal(""))
		})
		It("v2 serialized okay", func() {
			e := NewSwitchableAPIGatewayResponseV2(&events.APIGatewayV2HTTPResponse{})
			b, err := json.Marshal(e)
			Expect(err).To(BeNil())
			m := map[string]interface{}{}
			err = json.Unmarshal(b, &m)
			Expect(err).To(BeNil())
			Expect(m["statusCode"]).To(Equal(0.0))
			Expect(m["body"]).To(Equal(""))
		})
	})
	Context("Deserialization", func() {
		It("v1 deserialized okay", func() {
			input := &events.APIGatewayProxyResponse{
				StatusCode: 123,
				Body:       "234",
			}
			b, _ := json.Marshal(input)
			s := SwitchableAPIGatewayResponse{}
			err := s.UnmarshalJSON(b)
			Expect(err).To(BeNil())
			Expect(s.Version2()).To(BeNil())
			Expect(s.Version1()).To(BeEquivalentTo(input))
		})
		It("v2 deserialized okay", func() {
			input := &events.APIGatewayV2HTTPResponse{
				StatusCode: 123,
				Body:       "234",
				Cookies:    []string{"4", "5"},
			}
			b, _ := json.Marshal(input)
			s := SwitchableAPIGatewayResponse{}
			err := s.UnmarshalJSON(b)
			Expect(err).To(BeNil())
			Expect(s.Version1()).To(BeNil())
			Expect(s.Version2()).To(BeEquivalentTo(input))
		})
	})
})

