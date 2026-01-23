package bmc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBMC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BMC Suite")
}
