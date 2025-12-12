// pkg/bmc/manager.go
package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/starbops/taozhai/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// BMCJobGVR is the GroupVersionResource for BMC Jobs
	BMCJobGVR = schema.GroupVersionResource{
		Group:    "bmc.tinkerbell.org",
		Version:  "v1alpha1",
		Resource: "jobs",
	}
)

// Manager manages BMC Job resources
type Manager struct {
	client *client.Client
}

// NewManager creates a new BMC manager
func NewManager(c *client.Client) *Manager {
	return &Manager{client: c}
}

// CreatePowerOffJob creates a BMC Job to power off a server
// The machineName parameter is the name of both the Inventory CR and Machine CR (they share the same name)
func (m *Manager) CreatePowerOffJob(ctx context.Context, machineName, namespace string) (string, error) {
	fmt.Printf("\nCreating BMC power-off Job for machine: %s\n", machineName)

	// Machine CR name = Inventory CR name (same namespace)
	// Just create the Job referencing the Machine directly
	jobName := fmt.Sprintf("taozhai-poweroff-%s-%d", machineName, time.Now().Unix())
	job := m.createJobSpec(jobName, machineName, namespace)

	createdJob, err := m.client.DynamicClient.
		Resource(BMCJobGVR).
		Namespace(namespace).
		Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create BMC Job: %w", err)
	}

	jobName = createdJob.GetName()
	fmt.Printf("✓ BMC Job created: %s\n", jobName)
	return jobName, nil
}

// createJobSpec creates the BMC Job specification
// The Job references a Machine CR which has the same name as the Inventory CR
func (m *Manager) createJobSpec(jobName, machineName, machineNamespace string) *unstructured.Unstructured {
	job := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "bmc.tinkerbell.org/v1alpha1",
			"kind":       "Job",
			"metadata": map[string]interface{}{
				"name": jobName,
				"labels": map[string]interface{}{
					"app":     "taozhai",
					"machine": machineName,
					"action":  "power-off",
				},
			},
			"spec": map[string]interface{}{
				"machineRef": map[string]interface{}{
					"name":      machineName,
					"namespace": machineNamespace,
				},
				"tasks": []interface{}{
					map[string]interface{}{
						"powerAction": "off",
					},
				},
			},
		},
	}

	return job
}

// GetJobStatus retrieves the status of a BMC Job
func (m *Manager) GetJobStatus(ctx context.Context, jobName, namespace string) (string, error) {
	job, err := m.client.DynamicClient.
		Resource(BMCJobGVR).
		Namespace(namespace).
		Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get BMC Job: %w", err)
	}

	// Extract status
	status, found, err := unstructured.NestedString(job.Object, "status", "state")
	if err != nil || !found {
		return "unknown", nil
	}

	return status, nil
}

// DeleteJob deletes a BMC Job
func (m *Manager) DeleteJob(ctx context.Context, jobName, namespace string) error {
	err := m.client.DynamicClient.
		Resource(BMCJobGVR).
		Namespace(namespace).
		Delete(ctx, jobName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete BMC Job: %w", err)
	}

	fmt.Printf("✓ BMC Job deleted: %s\n", jobName)
	return nil
}
