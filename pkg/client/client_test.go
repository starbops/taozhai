package client_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/starbops/taozhai/pkg/client"
	"k8s.io/client-go/rest"
)

var _ = Describe("Client", func() {
	Describe("NewClient", func() {
		Context("with valid config", func() {
			It("should create a client with both Clientset and DynamicClient", func() {
				// Create a fake config that simulates an in-cluster config
				// Note: This will fail in unit tests without a real cluster,
				// but we can verify the function signature and error handling
				config := &rest.Config{
					Host: "https://localhost:6443",
				}

				c, err := client.NewClient(config)
				// In a real test environment without a cluster, this will succeed
				// since the fake config doesn't require actual connectivity for client creation
				Expect(err).NotTo(HaveOccurred())
				Expect(c).NotTo(BeNil())
				Expect(c.Clientset).NotTo(BeNil())
				Expect(c.DynamicClient).NotTo(BeNil())
			})
		})

		Context("with nil config", func() {
			It("should panic", func() {
				// NewClient panics when given nil config since kubernetes.NewForConfig
				// does not check for nil
				Expect(func() {
					_, _ = client.NewClient(nil)
				}).To(Panic())
			})
		})
	})
})
