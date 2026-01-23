package inventory_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/starbops/taozhai/internal/testutil"
	"github.com/starbops/taozhai/pkg/inventory"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Finder", func() {
	var (
		ctx    context.Context
		finder *inventory.Finder
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("NewFinder", func() {
		It("should create a finder", func() {
			client := testutil.NewFakeClient(nil)
			finder = inventory.NewFinder(client)
			Expect(finder).NotTo(BeNil())
		})
	})

	Describe("GetHardwareByMAC", func() {
		Context("when inventory exists with matching MAC", func() {
			BeforeEach(func() {
				inv := testutil.NewInventoryBuilder("test-server", "tink-system").
					WithMAC("aa:bb:cc:dd:ee:ff").
					WithPrimaryDisk("/dev/sda").
					WithArch("amd64").
					Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					inv,
				)
				finder = inventory.NewFinder(client)
			})

			It("should find the inventory by exact MAC match", func() {
				result, err := finder.GetHardwareByMAC(ctx, "aa:bb:cc:dd:ee:ff")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.GetName()).To(Equal("test-server"))
			})

			It("should find the inventory with normalized MAC (uppercase)", func() {
				result, err := finder.GetHardwareByMAC(ctx, "AA:BB:CC:DD:EE:FF")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.GetName()).To(Equal("test-server"))
			})

			It("should find the inventory with normalized MAC (dashes)", func() {
				result, err := finder.GetHardwareByMAC(ctx, "aa-bb-cc-dd-ee-ff")
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.GetName()).To(Equal("test-server"))
			})
		})

		Context("when no inventory matches", func() {
			BeforeEach(func() {
				inv := testutil.NewInventoryBuilder("test-server", "tink-system").
					WithMAC("aa:bb:cc:dd:ee:ff").
					Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					inv,
				)
				finder = inventory.NewFinder(client)
			})

			It("should return an error", func() {
				_, err := finder.GetHardwareByMAC(ctx, "11:22:33:44:55:66")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no inventory found"))
			})
		})

		Context("when multiple inventories exist", func() {
			BeforeEach(func() {
				inv1 := testutil.NewInventoryBuilder("server-1", "tink-system").
					WithMAC("aa:bb:cc:dd:ee:ff").
					Build()
				inv2 := testutil.NewInventoryBuilder("server-2", "tink-system").
					WithMAC("11:22:33:44:55:66").
					Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					inv1, inv2,
				)
				finder = inventory.NewFinder(client)
			})

			It("should find the correct inventory", func() {
				result, err := finder.GetHardwareByMAC(ctx, "11:22:33:44:55:66")
				Expect(err).NotTo(HaveOccurred())
				Expect(result.GetName()).To(Equal("server-2"))
			})
		})

		Context("when inventory has no MAC set", func() {
			BeforeEach(func() {
				inv := testutil.NewInventoryBuilder("test-server", "tink-system").
					Build() // No MAC set

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					inv,
				)
				finder = inventory.NewFinder(client)
			})

			It("should not match and return error", func() {
				_, err := finder.GetHardwareByMAC(ctx, "aa:bb:cc:dd:ee:ff")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no inventory found"))
			})
		})
	})

	Describe("FindByMAC", func() {
		Context("when inventory exists", func() {
			BeforeEach(func() {
				inv := testutil.NewInventoryBuilder("test-server", "tink-system").
					WithMAC("aa:bb:cc:dd:ee:ff").
					WithPrimaryDisk("/dev/sda").
					WithArch("amd64").
					WithBMCSpec("192.168.1.10", 443).
					Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					inv,
				)
				finder = inventory.NewFinder(client)
			})

			It("should return formatted output with all fields", func() {
				output, err := finder.FindByMAC(ctx, "aa:bb:cc:dd:ee:ff")
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("Name:      test-server"))
				Expect(output).To(ContainSubstring("Namespace: tink-system"))
				Expect(output).To(ContainSubstring("MAC:       aa:bb:cc:dd:ee:ff"))
				Expect(output).To(ContainSubstring("Disk:      /dev/sda"))
				Expect(output).To(ContainSubstring("Arch:      amd64"))
				Expect(output).To(ContainSubstring("BMC Configuration:"))
				Expect(output).To(ContainSubstring("Host: 192.168.1.10"))
				Expect(output).To(ContainSubstring("Port: 443"))
			})
		})
	})
})
