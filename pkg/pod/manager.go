// pkg/pod/manager.go
package pod

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/starbops/taozhai/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manager handles pod operations
type Manager struct {
	client *client.Client
}

// NewManager creates a new pod manager
func NewManager(c *client.Client) *Manager {
	return &Manager{client: c}
}

// Create creates a pod
func (m *Manager) Create(ctx context.Context, namespace string, pod *corev1.Pod) error {
	_, err := m.client.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	return err
}

// Delete deletes a pod
func (m *Manager) Delete(ctx context.Context, namespace, name string) error {
	fmt.Printf("Cleaning up pod %s...\n", name)
	return m.client.Clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// WaitForCompletion waits for a pod to complete
func (m *Manager) WaitForCompletion(ctx context.Context, namespace, name string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		pod, err := m.client.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		switch pod.Status.Phase {
		case corev1.PodSucceeded:
			return nil
		case corev1.PodFailed:
			return fmt.Errorf("pod failed")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			continue
		}
	}
}

// GetLogs retrieves logs from a pod
func (m *Manager) GetLogs(ctx context.Context, namespace, name string) (string, error) {
	req := m.client.Clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	buf := new(strings.Builder)
	if _, err := buf.ReadFrom(stream); err != nil {
		return "", err
	}

	return buf.String(), nil
}
