package httpadapter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHTTPAdapter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HttpAdapter Suite")
}
