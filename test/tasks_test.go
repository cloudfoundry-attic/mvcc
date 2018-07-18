package test_test

import (
	"context"

	"code.cloudfoundry.org/mvcc"
	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tasks", func() {
	Describe("GET /v3/tasks/:guid", func() {
		var (
			space mvcc.Space
			task  mvcc.Task
		)

		BeforeEach(func() {
			var err error

			org, err := cc.V3CreateOrganization(admin.AccessToken)
			Expect(err).NotTo(HaveOccurred())

			space, err = cc.V3CreateSpace(admin.AccessToken, org)
			Expect(err).NotTo(HaveOccurred())

			app, err := cc.V3CreateApp(admin.AccessToken, space)
			Expect(err).NotTo(HaveOccurred())

			task, err = cc.V3CreateTask(admin.AccessToken, app)
			Expect(err).NotTo(HaveOccurred())
		})

		It("succeeds when the subject has `task.read` for the parent space", func() {
			permission := perm.Permission{
				Action:          "task.read",
				ResourcePattern: space.UUID,
			}
			roleName := mvcc.RandomUUID("space-read-task")

			_, err := permClient.CreateRole(context.Background(), roleName, permission)
			Expect(err).NotTo(HaveOccurred())

			defer permClient.DeleteRole(context.Background(), roleName)

			err = permClient.AssignRole(context.Background(), roleName, actor)
			Expect(err).NotTo(HaveOccurred())

			t, err := cc.V3GetTask(user.AccessToken, task.UUID)
			Expect(err).NotTo(HaveOccurred())
			Expect(t).To(Equal(task))
		})

		It("fails when the subject has `task.read` for a different space", func() {
			permission := perm.Permission{
				Action:          "task.read",
				ResourcePattern: mvcc.RandomUUID("other-space"),
			}
			roleName := mvcc.RandomUUID("space-read-task")

			_, err := permClient.CreateRole(context.Background(), roleName, permission)
			Expect(err).NotTo(HaveOccurred())

			defer permClient.DeleteRole(context.Background(), roleName)

			err = permClient.AssignRole(context.Background(), roleName, actor)
			Expect(err).NotTo(HaveOccurred())

			_, err = cc.V3GetTask(user.AccessToken, task.UUID)
			Expect(err).To(MatchError(mvcc.ErrNotFound))
		})
	})
})
