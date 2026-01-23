// Package testutil provides testing utilities for taozhai
package testutil

import (
	"github.com/starbops/taozhai/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

// NewFakeClient creates a fake client for testing
func NewFakeClient(kubeObjects []runtime.Object, dynamicObjects ...runtime.Object) *client.Client {
	fakeClientset := kubefake.NewSimpleClientset(kubeObjects...)
	fakeDynamic := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), dynamicObjects...)

	return &client.Client{
		Clientset:     fakeClientset,
		DynamicClient: fakeDynamic,
	}
}

// NewFakeClientWithScheme creates a fake client with a custom scheme
func NewFakeClientWithScheme(scheme *runtime.Scheme, kubeObjects []runtime.Object, dynamicObjects ...runtime.Object) *client.Client {
	fakeClientset := kubefake.NewSimpleClientset(kubeObjects...)
	fakeDynamic := dynamicfake.NewSimpleDynamicClient(scheme, dynamicObjects...)

	return &client.Client{
		Clientset:     fakeClientset,
		DynamicClient: fakeDynamic,
	}
}
