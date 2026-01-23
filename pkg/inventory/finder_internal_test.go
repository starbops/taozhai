package inventory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This file tests unexported functions in the inventory package

var _ = Describe("normalizeMAC", func() {
	DescribeTable("MAC address normalization",
		func(input, expected string) {
			Expect(normalizeMAC(input)).To(Equal(expected))
		},
		Entry("lowercase with colons (passthrough)", "aa:bb:cc:dd:ee:ff", "aa:bb:cc:dd:ee:ff"),
		Entry("uppercase with colons", "AA:BB:CC:DD:EE:FF", "aa:bb:cc:dd:ee:ff"),
		Entry("mixed case with colons", "Aa:Bb:Cc:Dd:Ee:Ff", "aa:bb:cc:dd:ee:ff"),
		Entry("lowercase with dashes", "aa-bb-cc-dd-ee-ff", "aa:bb:cc:dd:ee:ff"),
		Entry("uppercase with dashes", "AA-BB-CC-DD-EE-FF", "aa:bb:cc:dd:ee:ff"),
		Entry("mixed case with dashes", "Aa-Bb-Cc-Dd-Ee-Ff", "aa:bb:cc:dd:ee:ff"),
		Entry("empty string", "", ""),
	)
})
