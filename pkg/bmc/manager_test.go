package bmc_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/starbops/taozhai/internal/testutil"
	"github.com/starbops/taozhai/pkg/bmc"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Manager", func() {
	var (
		ctx     context.Context
		manager *bmc.Manager
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("NewManager", func() {
		It("should create a manager", func() {
			client := testutil.NewFakeClient(nil)
			manager = bmc.NewManager(client)
			Expect(manager).NotTo(BeNil())
		})
	})

	Describe("CreatePowerOffJob", func() {
		BeforeEach(func() {
			client := testutil.NewFakeClientWithScheme(runtime.NewScheme(), nil)
			manager = bmc.NewManager(client)
		})

		It("should create a BMC Job with correct name format", func() {
			jobName, err := manager.CreatePowerOffJob(ctx, "test-machine", "tink-system")
			Expect(err).NotTo(HaveOccurred())
			Expect(jobName).To(ContainSubstring("taozhai-poweroff-test-machine"))
		})

		It("should create a job in the specified namespace", func() {
			jobName, err := manager.CreatePowerOffJob(ctx, "my-server", "custom-ns")
			Expect(err).NotTo(HaveOccurred())
			Expect(jobName).To(ContainSubstring("taozhai-poweroff-my-server"))
		})
	})

	Describe("GetJobStatus", func() {
		Context("when job has state set", func() {
			BeforeEach(func() {
				job := testutil.NewBMCJobBuilder("test-job", "tink-system").
					WithState("running").
					Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					job,
				)
				manager = bmc.NewManager(client)
			})

			It("should return the job state", func() {
				status, err := manager.GetJobStatus(ctx, "test-job", "tink-system")
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal("running"))
			})
		})

		Context("when job has no state", func() {
			BeforeEach(func() {
				job := testutil.NewBMCJobBuilder("test-job", "tink-system").
					Build() // No state set

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					job,
				)
				manager = bmc.NewManager(client)
			})

			It("should return unknown", func() {
				status, err := manager.GetJobStatus(ctx, "test-job", "tink-system")
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal("unknown"))
			})
		})

		Context("when job does not exist", func() {
			BeforeEach(func() {
				client := testutil.NewFakeClientWithScheme(runtime.NewScheme(), nil)
				manager = bmc.NewManager(client)
			})

			It("should return an error", func() {
				_, err := manager.GetJobStatus(ctx, "nonexistent", "tink-system")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("DeleteJob", func() {
		Context("when job exists", func() {
			BeforeEach(func() {
				job := testutil.NewBMCJobBuilder("to-delete", "tink-system").Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					job,
				)
				manager = bmc.NewManager(client)
			})

			It("should delete the job", func() {
				err := manager.DeleteJob(ctx, "to-delete", "tink-system")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when job does not exist", func() {
			BeforeEach(func() {
				client := testutil.NewFakeClientWithScheme(runtime.NewScheme(), nil)
				manager = bmc.NewManager(client)
			})

			It("should return an error", func() {
				err := manager.DeleteJob(ctx, "nonexistent", "tink-system")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("WaitForJobCompletion", func() {
		Context("when job completes successfully", func() {
			BeforeEach(func() {
				job := testutil.NewBMCJobBuilder("completed-job", "tink-system").
					WithCompletedCondition().
					Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					job,
				)
				manager = bmc.NewManager(client)
			})

			It("should return nil", func() {
				err := manager.WaitForJobCompletion(ctx, "completed-job", "tink-system", 5*time.Second)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when job fails", func() {
			BeforeEach(func() {
				job := testutil.NewBMCJobBuilder("failed-job", "tink-system").
					WithFailedCondition().
					Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					job,
				)
				manager = bmc.NewManager(client)
			})

			It("should return an error", func() {
				err := manager.WaitForJobCompletion(ctx, "failed-job", "tink-system", 5*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("BMC Job failed"))
			})
		})

		Context("when job does not exist", func() {
			BeforeEach(func() {
				client := testutil.NewFakeClientWithScheme(runtime.NewScheme(), nil)
				manager = bmc.NewManager(client)
			})

			It("should return an error", func() {
				err := manager.WaitForJobCompletion(ctx, "nonexistent", "tink-system", 5*time.Second)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when timeout occurs with no conditions", func() {
			BeforeEach(func() {
				// Job with no conditions set
				job := testutil.NewBMCJobBuilder("pending-job", "tink-system").Build()

				client := testutil.NewFakeClientWithScheme(
					runtime.NewScheme(),
					nil,
					job,
				)
				manager = bmc.NewManager(client)
			})

			It("should return timeout error", func() {
				err := manager.WaitForJobCompletion(ctx, "pending-job", "tink-system", 100*time.Millisecond)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("timeout"))
			})
		})
	})
})
