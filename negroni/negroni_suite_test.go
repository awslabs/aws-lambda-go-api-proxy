package negroniadapter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNegroni(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NegroniAdapter Suite")
}
