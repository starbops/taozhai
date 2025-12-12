// cmd/taozhai/main.go
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "path to kubeconfig file")
	}
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: taozhai [--kubeconfig PATH] <target-ip>\n")
		os.Exit(1)
	}
	targetIP := args[0]

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating kubernetes client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Create discovery pod
	podName := fmt.Sprintf("taozhai-discovery-%d", time.Now().Unix())
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "discovery",
					Image:   "alpine:latest",
					Command: []string{"/bin/sh", "-c"},
					Args: []string{
						fmt.Sprintf("ping -c 3 %s; ip neigh show %s", targetIP, targetIP),
					},
				},
			},
			HostNetwork: true,
		},
	}

	fmt.Printf("Creating discovery pod %s...\n", podName)
	_, err = clientset.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pod: %v\n", err)
		os.Exit(1)
	}

	// Wait for pod to complete
	fmt.Println("Waiting for pod to complete...")
	if err := waitForPodCompletion(ctx, clientset, "default", podName, 2*time.Minute); err != nil {
		fmt.Fprintf(os.Stderr, "Error waiting for pod: %v\n", err)
		cleanupPod(ctx, clientset, "default", podName)
		os.Exit(1)
	}

	// Get pod logs
	logs, err := getPodLogs(ctx, clientset, "default", podName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting pod logs: %v\n", err)
		cleanupPod(ctx, clientset, "default", podName)
		os.Exit(1)
	}

	// Extract MAC address from logs
	macAddr := extractMACAddress(logs)
	if macAddr == "" {
		fmt.Fprintf(os.Stderr, "Could not extract MAC address from logs:\n%s\n", logs)
		cleanupPod(ctx, clientset, "default", podName)
		os.Exit(1)
	}

	fmt.Printf("Found MAC address: %s\n", macAddr)

	// Cleanup discovery pod
	cleanupPod(ctx, clientset, "default", podName)

	// Search for inventory
	fmt.Printf("\nSearching for Inventory with MAC %s...\n", macAddr)
	inventory, err := findInventory(ctx, clientset, *kubeconfig, macAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding inventory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + inventory)
}

func waitForPodCompletion(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			if pod.Status.Phase == corev1.PodFailed {
				return fmt.Errorf("pod failed")
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func getPodLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (string, error) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{})
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	return buf.String(), err
}

func extractMACAddress(logs string) string {
	// Match MAC address pattern (e.g., aa:bb:cc:dd:ee:ff)
	macRegex := regexp.MustCompile(`([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}`)
	match := macRegex.FindString(logs)
	return strings.ToLower(match)
}

func cleanupPod(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) {
	fmt.Printf("Cleaning up pod %s...\n", name)
	clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func findInventory(ctx context.Context, clientset *kubernetes.Clientset, kubeconfig, macAddr string) (string, error) {
	// Use kubectl command to search inventories
	// This is a simplified version - you'd want to use dynamic client or CRD client
	cmd := fmt.Sprintf("kubectl --kubeconfig=%s -n tink-system get inventories -o yaml | grep -i %s", kubeconfig, macAddr)

	// For now, return instruction to run manually
	// In production, use exec.Command or dynamic client
	return fmt.Sprintf("Run: %s", cmd), nil
}
