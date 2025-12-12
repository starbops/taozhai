// pkg/inventory/finder.go
package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/starbops/taozhai/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	TinkSystemNamespace = "tink-system"
)

var (
	InventoryGVR = schema.GroupVersionResource{
		Group:    "tinkerbell.org",
		Version:  "v1alpha1",
		Resource: "hardware",
	}
)

// Finder finds inventory resources
type Finder struct {
	client *client.Client
}

// NewFinder creates a new inventory finder
func NewFinder(c *client.Client) *Finder {
	return &Finder{client: c}
}

// FindByMAC finds an inventory by MAC address
func (f *Finder) FindByMAC(ctx context.Context, macAddr string) (string, error) {
	fmt.Printf("\nSearching for Inventory with MAC %s...\n", macAddr)

	list, err := f.client.DynamicClient.
		Resource(InventoryGVR).
		Namespace(TinkSystemNamespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return "", fmt.Errorf("failed to list inventories: %w", err)
	}

	normalizedMAC := normalizeMAC(macAddr)

	for _, item := range list.Items {
		if f.containsMAC(&item, normalizedMAC) {
			return f.formatInventory(&item), nil
		}
	}

	return "", fmt.Errorf("no inventory found with MAC address %s", macAddr)
}

// containsMAC checks if the inventory contains the MAC address
func (f *Finder) containsMAC(inventory *unstructured.Unstructured, macAddr string) bool {
	// Check in spec.interfaces[].dhcp.mac or similar fields
	interfaces, found, err := unstructured.NestedSlice(inventory.Object, "spec", "interfaces")
	if err != nil || !found {
		return false
	}

	for _, iface := range interfaces {
		ifaceMap, ok := iface.(map[string]interface{})
		if !ok {
			continue
		}

		// Check dhcp.mac
		if dhcp, found, _ := unstructured.NestedMap(ifaceMap, "dhcp"); found {
			if mac, ok := dhcp["mac"].(string); ok && normalizeMAC(mac) == macAddr {
				return true
			}
		}

		// Check netboot.allowPXE (some CRDs store MAC differently)
		if netboot, found, _ := unstructured.NestedMap(ifaceMap, "netboot"); found {
			if mac, ok := netboot["mac"].(string); ok && normalizeMAC(mac) == macAddr {
				return true
			}
		}
	}

	return false
}

// formatInventory formats an inventory for display
func (f *Finder) formatInventory(inventory *unstructured.Unstructured) string {
	name := inventory.GetName()
	namespace := inventory.GetNamespace()

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Name:      %s\n", name))
	output.WriteString(fmt.Sprintf("Namespace: %s\n", namespace))

	// Extract and display relevant fields
	if metadata, found, _ := unstructured.NestedMap(inventory.Object, "spec", "metadata"); found {
		output.WriteString("\nMetadata:\n")
		for k, v := range metadata {
			output.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	if interfaces, found, _ := unstructured.NestedSlice(inventory.Object, "spec", "interfaces"); found {
		output.WriteString("\nInterfaces:\n")
		for i, iface := range interfaces {
			ifaceMap, ok := iface.(map[string]interface{})
			if !ok {
				continue
			}
			output.WriteString(fmt.Sprintf("  [%d]:\n", i))
			if dhcp, found, _ := unstructured.NestedMap(ifaceMap, "dhcp"); found {
				if mac, ok := dhcp["mac"].(string); ok {
					output.WriteString(fmt.Sprintf("    MAC: %s\n", mac))
				}
				if ip, ok := dhcp["ip"].(string); ok {
					output.WriteString(fmt.Sprintf("    IP:  %s\n", ip))
				}
			}
		}
	}

	return output.String()
}

// normalizeMAC normalizes MAC address format
func normalizeMAC(mac string) string {
	return strings.ToLower(strings.ReplaceAll(mac, "-", ":"))
}
