// pkg/discovery/discoverer.go
package discovery

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/starbops/taozhai/pkg/client"
	"github.com/starbops/taozhai/pkg/pod"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultNamespace  = "default"
	DefaultTimeout    = 2 * time.Minute
	DiscoveryImage    = "alpine:latest"
	MacvlanAnnotation = "k8s.v1.cni.cncf.io/networks"
)

// Discoverer handles MAC address discovery
type Discoverer struct {
	client       *client.Client
	podManager   *pod.Manager
	namespace    string
	timeout      time.Duration
	macExtractor *MACExtractor
}

// NewDiscoverer creates a new discoverer
func NewDiscoverer(c *client.Client) *Discoverer {
	return &Discoverer{
		client:       c,
		podManager:   pod.NewManager(c),
		namespace:    DefaultNamespace,
		timeout:      DefaultTimeout,
		macExtractor: NewMACExtractor(),
	}
}

// DiscoverMACFromIP discovers the MAC address for a given IP
func (d *Discoverer) DiscoverMACFromIP(ctx context.Context, targetIP string) (string, error) {
	// Validate IP address to prevent command injection
	if net.ParseIP(targetIP) == nil {
		return "", fmt.Errorf("invalid IP address: %s", targetIP)
	}

	podName := fmt.Sprintf("taozhai-discovery-%d", time.Now().Unix())

	podSpec := d.buildDiscoveryPodSpec(podName, targetIP)

	fmt.Printf("Creating discovery pod %s...\n", podName)
	if err := d.podManager.Create(ctx, d.namespace, podSpec); err != nil {
		return "", fmt.Errorf("failed to create pod: %w", err)
	}
	defer d.podManager.Delete(ctx, d.namespace, podName)

	fmt.Println("Waiting for pod to complete...")
	if err := d.podManager.WaitForCompletion(ctx, d.namespace, podName, d.timeout); err != nil {
		return "", fmt.Errorf("pod did not complete successfully: %w", err)
	}

	logs, err := d.podManager.GetLogs(ctx, d.namespace, podName)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	macAddr, err := d.macExtractor.Extract(logs)
	if err != nil {
		return "", fmt.Errorf("failed to extract MAC address: %w\nLogs:\n%s", err, logs)
	}

	return macAddr, nil
}

func (d *Discoverer) buildDiscoveryPodSpec(name, targetIP string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: d.namespace,
			Annotations: map[string]string{
				MacvlanAnnotation: "macvlan",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "discovery",
					Image:   DiscoveryImage,
					Command: []string{"/bin/sh", "-c"},
					Args: []string{
						fmt.Sprintf("ping -c 3 %s; ip neigh show %s", targetIP, targetIP),
					},
				},
			},
			HostNetwork: true,
		},
	}
}

// MACExtractor extracts MAC addresses from text
type MACExtractor struct {
	regex *regexp.Regexp
}

// NewMACExtractor creates a new MAC extractor
func NewMACExtractor() *MACExtractor {
	return &MACExtractor{
		regex: regexp.MustCompile(`([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}`),
	}
}

// Extract finds and returns the first MAC address in the text
func (m *MACExtractor) Extract(text string) (string, error) {
	match := m.regex.FindString(text)
	if match == "" {
		return "", fmt.Errorf("no MAC address found")
	}
	return strings.ToLower(match), nil
}
