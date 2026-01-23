package testutil

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PodBuilder helps construct test pods
type PodBuilder struct {
	pod *corev1.Pod
}

// NewPodBuilder creates a new PodBuilder
func NewPodBuilder(name, namespace string) *PodBuilder {
	return &PodBuilder{
		pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "test", Image: "alpine"},
				},
			},
		},
	}
}

// WithPhase sets the pod phase
func (b *PodBuilder) WithPhase(phase corev1.PodPhase) *PodBuilder {
	b.pod.Status.Phase = phase
	return b
}

// WithLabels sets the pod labels
func (b *PodBuilder) WithLabels(labels map[string]string) *PodBuilder {
	b.pod.Labels = labels
	return b
}

// Build returns the constructed pod
func (b *PodBuilder) Build() *corev1.Pod {
	return b.pod
}

// InventoryBuilder helps construct test Inventory CRs
type InventoryBuilder struct {
	obj *unstructured.Unstructured
}

// NewInventoryBuilder creates a new InventoryBuilder
func NewInventoryBuilder(name, namespace string) *InventoryBuilder {
	return &InventoryBuilder{
		obj: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "metal.harvesterhci.io/v1alpha1",
				"kind":       "Inventory",
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": namespace,
				},
				"spec": map[string]interface{}{},
			},
		},
	}
}

// WithMAC sets the management interface MAC address
func (b *InventoryBuilder) WithMAC(mac string) *InventoryBuilder {
	spec := b.obj.Object["spec"].(map[string]interface{})
	spec["managementInterfaceMacAddress"] = mac
	return b
}

// WithPrimaryDisk sets the primary disk
func (b *InventoryBuilder) WithPrimaryDisk(disk string) *InventoryBuilder {
	spec := b.obj.Object["spec"].(map[string]interface{})
	spec["primaryDisk"] = disk
	return b
}

// WithArch sets the architecture
func (b *InventoryBuilder) WithArch(arch string) *InventoryBuilder {
	spec := b.obj.Object["spec"].(map[string]interface{})
	spec["arch"] = arch
	return b
}

// WithBMCSpec sets the BMC spec
func (b *InventoryBuilder) WithBMCSpec(host string, port int) *InventoryBuilder {
	spec := b.obj.Object["spec"].(map[string]interface{})
	spec["baseboardManagementSpec"] = map[string]interface{}{
		"connection": map[string]interface{}{
			"host": host,
			"port": float64(port),
		},
	}
	return b
}

// Build returns the constructed inventory
func (b *InventoryBuilder) Build() *unstructured.Unstructured {
	return b.obj
}

// BMCJobBuilder helps construct test BMC Job CRs
type BMCJobBuilder struct {
	obj *unstructured.Unstructured
}

// NewBMCJobBuilder creates a new BMCJobBuilder
func NewBMCJobBuilder(name, namespace string) *BMCJobBuilder {
	return &BMCJobBuilder{
		obj: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "bmc.tinkerbell.org/v1alpha1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": namespace,
				},
				"spec":   map[string]interface{}{},
				"status": map[string]interface{}{},
			},
		},
	}
}

// WithMachineRef sets the machine reference
func (b *BMCJobBuilder) WithMachineRef(name, namespace string) *BMCJobBuilder {
	spec := b.obj.Object["spec"].(map[string]interface{})
	spec["machineRef"] = map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	}
	return b
}

// WithCompletedCondition adds a completed condition
func (b *BMCJobBuilder) WithCompletedCondition() *BMCJobBuilder {
	b.obj.Object["status"] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "Completed",
				"status": "True",
			},
		},
	}
	return b
}

// WithFailedCondition adds a failed condition
func (b *BMCJobBuilder) WithFailedCondition() *BMCJobBuilder {
	b.obj.Object["status"] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "Failed",
				"status": "True",
			},
		},
	}
	return b
}

// WithState sets the status state
func (b *BMCJobBuilder) WithState(state string) *BMCJobBuilder {
	status := b.obj.Object["status"].(map[string]interface{})
	status["state"] = state
	return b
}

// Build returns the constructed BMC Job
func (b *BMCJobBuilder) Build() *unstructured.Unstructured {
	return b.obj
}
