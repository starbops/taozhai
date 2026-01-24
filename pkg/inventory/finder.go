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
	// InventoryGVR is the GroupVersionResource for Seeder Inventory CRs
	//
	// API Stability: v1alpha1 (unstable)
	//   - Breaking changes may occur in future versions
	//   - Monitor upstream Harvester/Seeder for API graduation
	//   - Upstream: https://github.com/harvester/seeder
	InventoryGVR = schema.GroupVersionResource{
		Group:    "metal.harvesterhci.io",
		Version:  "v1alpha1",
		Resource: "inventories",
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

// FindByMAC finds an inventory by MAC address and returns formatted output
func (f *Finder) FindByMAC(ctx context.Context, macAddr string) (string, error) {
	fmt.Printf("\nSearching for Inventory with MAC %s...\n", macAddr)

	hardware, err := f.GetHardwareByMAC(ctx, macAddr)
	if err != nil {
		return "", err
	}

	return f.formatInventory(hardware), nil
}

// GetHardwareByMAC finds a Hardware CR by MAC address and returns the CR object
func (f *Finder) GetHardwareByMAC(ctx context.Context, macAddr string) (*unstructured.Unstructured, error) {
	list, err := f.client.DynamicClient.
		Resource(InventoryGVR).
		Namespace(TinkSystemNamespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to list inventories: %w", err)
	}

	normalizedMAC := normalizeMAC(macAddr)

	for _, item := range list.Items {
		if f.containsMAC(&item, normalizedMAC) {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("no inventory found with MAC address %s", macAddr)
}

// containsMAC checks if the inventory contains the MAC address
func (f *Finder) containsMAC(inventory *unstructured.Unstructured, macAddr string) bool {
	// Check .spec.managementInterfaceMacAddress for Seeder Inventory CRs
	mgmtMAC, found, err := unstructured.NestedString(
		inventory.Object,
		"spec",
		"managementInterfaceMacAddress",
	)
	if err != nil || !found {
		return false
	}

	return normalizeMAC(mgmtMAC) == macAddr
}

// formatInventory formats an inventory for display
func (f *Finder) formatInventory(inventory *unstructured.Unstructured) string {
	name := inventory.GetName()
	namespace := inventory.GetNamespace()

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Name:      %s\n", name))
	output.WriteString(fmt.Sprintf("Namespace: %s\n", namespace))

	// Extract and display Seeder Inventory fields
	if mgmtMAC, found, _ := unstructured.NestedString(inventory.Object, "spec", "managementInterfaceMacAddress"); found {
		output.WriteString(fmt.Sprintf("MAC:       %s\n", mgmtMAC))
	}

	if primaryDisk, found, _ := unstructured.NestedString(inventory.Object, "spec", "primaryDisk"); found {
		output.WriteString(fmt.Sprintf("Disk:      %s\n", primaryDisk))
	}

	if arch, found, _ := unstructured.NestedString(inventory.Object, "spec", "arch"); found {
		output.WriteString(fmt.Sprintf("Arch:      %s\n", arch))
	}

	// Show BMC spec if available
	if bmcSpec, found, _ := unstructured.NestedMap(inventory.Object, "spec", "baseboardManagementSpec"); found {
		output.WriteString("\nBMC Configuration:\n")
		if connection, ok := bmcSpec["connection"].(map[string]interface{}); ok {
			if host, ok := connection["host"].(string); ok {
				output.WriteString(fmt.Sprintf("  Host: %s\n", host))
			}
			if port, ok := connection["port"].(float64); ok {
				output.WriteString(fmt.Sprintf("  Port: %.0f\n", port))
			}
		}
	}

	return output.String()
}

// normalizeMAC normalizes MAC address format
func normalizeMAC(mac string) string {
	return strings.ToLower(strings.ReplaceAll(mac, "-", ":"))
}
