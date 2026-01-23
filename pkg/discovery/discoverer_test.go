package discovery_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/starbops/taozhai/pkg/discovery"
)

var _ = Describe("MACExtractor", func() {
	var extractor *discovery.MACExtractor

	BeforeEach(func() {
		extractor = discovery.NewMACExtractor()
	})

	Describe("Extract", func() {
		Context("when text contains a valid MAC address", func() {
			It("should extract MAC from standard ARP output", func() {
				text := "192.168.1.100 dev eth0 lladdr aa:bb:cc:dd:ee:ff REACHABLE"
				mac, err := extractor.Extract(text)
				Expect(err).NotTo(HaveOccurred())
				Expect(mac).To(Equal("aa:bb:cc:dd:ee:ff"))
			})

			It("should extract MAC and normalize uppercase to lowercase", func() {
				text := "192.168.1.100 dev eth0 lladdr AA:BB:CC:DD:EE:FF REACHABLE"
				mac, err := extractor.Extract(text)
				Expect(err).NotTo(HaveOccurred())
				Expect(mac).To(Equal("aa:bb:cc:dd:ee:ff"))
			})

			It("should extract MAC from multiline output", func() {
				text := `PING 192.168.1.100 (192.168.1.100): 56 data bytes
64 bytes from 192.168.1.100: icmp_seq=0 ttl=64 time=0.123 ms
64 bytes from 192.168.1.100: icmp_seq=1 ttl=64 time=0.089 ms
64 bytes from 192.168.1.100: icmp_seq=2 ttl=64 time=0.095 ms

--- 192.168.1.100 ping statistics ---
3 packets transmitted, 3 packets received, 0% packet loss
round-trip min/avg/max/stddev = 0.089/0.102/0.123/0.015 ms
192.168.1.100 dev eth0 lladdr 52:54:00:0b:2e:01 REACHABLE`
				mac, err := extractor.Extract(text)
				Expect(err).NotTo(HaveOccurred())
				Expect(mac).To(Equal("52:54:00:0b:2e:01"))
			})

			It("should extract the first MAC if multiple MACs are present", func() {
				text := "aa:bb:cc:dd:ee:ff and 11:22:33:44:55:66"
				mac, err := extractor.Extract(text)
				Expect(err).NotTo(HaveOccurred())
				Expect(mac).To(Equal("aa:bb:cc:dd:ee:ff"))
			})

			It("should handle mixed case MAC addresses", func() {
				text := "MAC: aA:Bb:Cc:dD:eE:Ff"
				mac, err := extractor.Extract(text)
				Expect(err).NotTo(HaveOccurred())
				Expect(mac).To(Equal("aa:bb:cc:dd:ee:ff"))
			})
		})

		Context("when text does not contain a MAC address", func() {
			It("should return an error for empty string", func() {
				_, err := extractor.Extract("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no MAC address found"))
			})

			It("should return an error for text without MAC", func() {
				text := "No MAC address here, just some random text"
				_, err := extractor.Extract(text)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no MAC address found"))
			})

			It("should not match invalid MAC formats", func() {
				text := "192.168.1.100 is not a MAC address"
				_, err := extractor.Extract(text)
				Expect(err).To(HaveOccurred())
			})

			It("should not match partial MAC addresses", func() {
				text := "aa:bb:cc:dd:ee is incomplete"
				_, err := extractor.Extract(text)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var _ = Describe("Discoverer", func() {
	Describe("NewDiscoverer", func() {
		It("should create a discoverer with default values", func() {
			d := discovery.NewDiscoverer(nil)
			Expect(d).NotTo(BeNil())
		})
	})

	Describe("SetTimeout", func() {
		It("should return the discoverer for chaining", func() {
			d := discovery.NewDiscoverer(nil)
			result := d.SetTimeout(60)
			Expect(result).To(Equal(d))
		})
	})

	Describe("SetNamespace", func() {
		It("should return the discoverer for chaining", func() {
			d := discovery.NewDiscoverer(nil)
			result := d.SetNamespace("test-ns")
			Expect(result).To(Equal(d))
		})
	})
})
