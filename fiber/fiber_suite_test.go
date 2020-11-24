package fiberadapter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFiber(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fiber Suite")
}
