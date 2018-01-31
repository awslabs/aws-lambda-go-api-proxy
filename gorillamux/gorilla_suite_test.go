package gorillamux_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gorilla Mux Suite")
}
