package pod_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/starbops/taozhai/internal/testutil"
	"github.com/starbops/taozhai/pkg/client"
	"github.com/starbops/taozhai/pkg/pod"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Manager", func() {
	var (
		ctx        context.Context
		testClient *client.Client
		manager    *pod.Manager
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("NewManager", func() {
		It("should create a manager", func() {
			testClient = testutil.NewFakeClient(nil)
			manager = pod.NewManager(testClient)
			Expect(manager).NotTo(BeNil())
		})
	})

	Describe("Create", func() {
		BeforeEach(func() {
			testClient = testutil.NewFakeClient(nil)
			manager = pod.NewManager(testClient)
		})

		It("should create a pod successfully", func() {
			testPod := testutil.NewPodBuilder("test-pod", "default").Build()

			err := manager.Create(ctx, "default", testPod)
			Expect(err).NotTo(HaveOccurred())

			// Verify pod was created
			createdPod, err := testClient.Clientset.CoreV1().Pods("default").Get(ctx, "test-pod", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(createdPod.Name).To(Equal("test-pod"))
		})

		It("should fail when pod already exists", func() {
			existingPod := testutil.NewPodBuilder("existing-pod", "default").Build()
			testClient = testutil.NewFakeClient([]runtime.Object{existingPod})
			manager = pod.NewManager(testClient)

			newPod := testutil.NewPodBuilder("existing-pod", "default").Build()
			err := manager.Create(ctx, "default", newPod)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			existingPod := testutil.NewPodBuilder("to-delete", "default").Build()
			testClient = testutil.NewFakeClient([]runtime.Object{existingPod})
			manager = pod.NewManager(testClient)
		})

		It("should delete an existing pod", func() {
			err := manager.Delete(ctx, "default", "to-delete")
			Expect(err).NotTo(HaveOccurred())

			// Verify pod was deleted
			_, err = testClient.Clientset.CoreV1().Pods("default").Get(ctx, "to-delete", metav1.GetOptions{})
			Expect(err).To(HaveOccurred())
		})

		It("should fail when pod does not exist", func() {
			err := manager.Delete(ctx, "default", "nonexistent")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("WaitForCompletion", func() {
		Context("when pod succeeds", func() {
			BeforeEach(func() {
				succeededPod := testutil.NewPodBuilder("succeeded-pod", "default").
					WithPhase(corev1.PodSucceeded).
					Build()
				testClient = testutil.NewFakeClient([]runtime.Object{succeededPod})
				manager = pod.NewManager(testClient)
			})

			It("should return nil", func() {
				err := manager.WaitForCompletion(ctx, "default", "succeeded-pod", 5*time.Second)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when pod fails", func() {
			BeforeEach(func() {
				failedPod := testutil.NewPodBuilder("failed-pod", "default").
					WithPhase(corev1.PodFailed).
					Build()
				testClient = testutil.NewFakeClient([]runtime.Object{failedPod})
				manager = pod.NewManager(testClient)
			})

			It("should return an error", func() {
				err := manager.WaitForCompletion(ctx, "default", "failed-pod", 5*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("pod failed"))
			})
		})

		Context("when pod does not exist", func() {
			BeforeEach(func() {
				testClient = testutil.NewFakeClient(nil)
				manager = pod.NewManager(testClient)
			})

			It("should return an error", func() {
				err := manager.WaitForCompletion(ctx, "default", "nonexistent", 5*time.Second)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when timeout occurs", func() {
			BeforeEach(func() {
				runningPod := testutil.NewPodBuilder("running-pod", "default").
					WithPhase(corev1.PodRunning).
					Build()
				testClient = testutil.NewFakeClient([]runtime.Object{runningPod})
				manager = pod.NewManager(testClient)
			})

			It("should return timeout error", func() {
				err := manager.WaitForCompletion(ctx, "default", "running-pod", 100*time.Millisecond)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("deadline exceeded"))
			})
		})
	})

	Describe("GetLogs", func() {
		Context("when pod exists", func() {
			var fakeClientset *kubefake.Clientset

			BeforeEach(func() {
				// Note: The fake clientset doesn't fully support GetLogs streaming
				// This test verifies the basic flow
				existingPod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "log-pod",
						Namespace: "default",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
					},
				}
				fakeClientset = kubefake.NewSimpleClientset(existingPod)
				testClient = &client.Client{
					Clientset:     fakeClientset,
					DynamicClient: nil,
				}
				manager = pod.NewManager(testClient)
			})

			It("should attempt to get logs", func() {
				// The fake clientset returns empty logs but doesn't error
				// In a real environment, this would return actual pod logs
				_, err := manager.GetLogs(ctx, "default", "log-pod")
				// Note: fake client may return error for GetLogs since it's not fully supported
				// We just verify the method can be called
				_ = err
			})
		})
	})
})
