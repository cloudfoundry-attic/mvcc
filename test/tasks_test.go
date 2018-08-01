package test_test

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/mvcc"
	. "code.cloudfoundry.org/mvcc/helpers"
	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tasks", func() {
	var (
		app   mvcc.App
		space mvcc.Space
		org   mvcc.Organization
		task  mvcc.Task

		dropletUUID string
	)

	BeforeEach(func() {
		var err error

		org, err = cc.V3CreateOrganization(admin.AccessToken)
		Expect(err).NotTo(HaveOccurred())

		space, err = cc.V3CreateSpace(admin.AccessToken, org)
		Expect(err).NotTo(HaveOccurred())

		app, err = cc.V3CreateApp(admin.AccessToken, space)
		Expect(err).NotTo(HaveOccurred())

		pkg, err := cc.V3CreatePackage(admin.AccessToken, app)
		Expect(err).NotTo(HaveOccurred())

		build, err := cc.V3CreateBuild(admin.AccessToken, pkg)
		Expect(err).NotTo(HaveOccurred())

		if err != nil {
			pkg, err = cc.V3GetPackage(admin.AccessToken, pkg.UUID)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println("pkg:", pkg, pkg.State)
		}
		Expect(err).NotTo(HaveOccurred())

		timer := time.NewTimer(time.Second * 5)
		ticker := time.NewTicker(time.Millisecond * 100)

		for _ = range ticker.C {
			build, err = cc.V3GetBuild(admin.AccessToken, build.UUID)
			Expect(err).NotTo(HaveOccurred())
			Expect(build.State).NotTo(Equal("FAILED"))

			if build.State == "STAGED" {
				dropletUUID = build.DropletUUID
				break
			} else if build.State == "FAILED" {
				fmt.Println("not staged", build.State, build)
			}

			Consistently(timer.C).ShouldNot(Receive())
		}

		task, err = cc.V3CreateTask(admin.AccessToken, app, dropletUUID)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := cc.V2DeleteOrganization(admin.AccessToken, org.UUID)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("GET /v3/tasks/:guid", func() {
		It("succeeds when the subject has `task.read` for the parent space", func() {
			permission := perm.Permission{
				Action:          "task.read",
				ResourcePattern: SpaceResourceID(org.UUID, space.UUID),
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

		It("succeeds when the subject has `task.read` for the parent org", func() {
			permission := perm.Permission{
				Action:          "task.read",
				ResourcePattern: OrgResourceID(org.UUID),
			}
			roleName := mvcc.RandomUUID("org-read-task")

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
				ResourcePattern: SpaceResourceID(org.UUID, mvcc.RandomUUID("other-space")),
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

		It("fails when the subject has `task.read` for a different organization", func() {
			permission := perm.Permission{
				Action:          "task.read",
				ResourcePattern: OrgResourceID(mvcc.RandomUUID("other-org")),
			}
			roleName := mvcc.RandomUUID("org-read-task")

			_, err := permClient.CreateRole(context.Background(), roleName, permission)
			Expect(err).NotTo(HaveOccurred())

			defer permClient.DeleteRole(context.Background(), roleName)

			err = permClient.AssignRole(context.Background(), roleName, actor)
			Expect(err).NotTo(HaveOccurred())

			_, err = cc.V3GetTask(user.AccessToken, task.UUID)
			Expect(err).To(MatchError(mvcc.ErrNotFound))
		})
	})

	Describe("GET /v3/tasks", func() {
		var (
			anotherTask mvcc.Task
			err         error
		)

		Context("when there are multiple tasks belonging to a space", func() {
			BeforeEach(func() {
				anotherTask, err = cc.V3CreateTask(admin.AccessToken, app, dropletUUID)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns both tasks when the subject has `task.read` for the parent space", func() {
				permission := perm.Permission{
					Action:          "task.read",
					ResourcePattern: SpaceResourceID(org.UUID, space.UUID),
				}
				roleName := mvcc.RandomUUID("space-read-task")

				_, err := permClient.CreateRole(context.Background(), roleName, permission)
				Expect(err).NotTo(HaveOccurred())

				defer permClient.DeleteRole(context.Background(), roleName)

				err = permClient.AssignRole(context.Background(), roleName, actor)
				Expect(err).NotTo(HaveOccurred())

				tasks, err := cc.V3ListTasks(user.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(HaveLen(2))
				Expect(tasks).To(ContainElement(task))
				Expect(tasks).To(ContainElement(anotherTask))
			})

			It("returns no tasks when the subject has `task.read` for another space", func() {
				permission := perm.Permission{
					Action:          "task.read",
					ResourcePattern: SpaceResourceID(org.UUID, "some-other-space"),
				}
				roleName := mvcc.RandomUUID("space-read-task")

				_, err := permClient.CreateRole(context.Background(), roleName, permission)
				Expect(err).NotTo(HaveOccurred())

				defer permClient.DeleteRole(context.Background(), roleName)

				err = permClient.AssignRole(context.Background(), roleName, actor)
				Expect(err).NotTo(HaveOccurred())

				tasks, err := cc.V3ListTasks(user.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(HaveLen(0))
			})
		})

		Context("when there are multiple tasks belonging to an organization", func() {
			BeforeEach(func() {
				anotherTask, err = cc.V3CreateTask(admin.AccessToken, app, dropletUUID)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns both tasks when the subject has `task.read` for the parent org", func() {
				permission := perm.Permission{
					Action:          "task.read",
					ResourcePattern: OrgResourceID(org.UUID),
				}
				roleName := mvcc.RandomUUID("org-read-task")

				_, err := permClient.CreateRole(context.Background(), roleName, permission)
				Expect(err).NotTo(HaveOccurred())

				defer permClient.DeleteRole(context.Background(), roleName)

				err = permClient.AssignRole(context.Background(), roleName, actor)
				Expect(err).NotTo(HaveOccurred())

				tasks, err := cc.V3ListTasks(user.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(HaveLen(2))
				Expect(tasks).To(ContainElement(task))
				Expect(tasks).To(ContainElement(anotherTask))
			})

			It("returns no tasks when the subject has `task.read` for another org", func() {
				permission := perm.Permission{
					Action:          "task.read",
					ResourcePattern: OrgResourceID("some-other-org"),
				}
				roleName := mvcc.RandomUUID("org-read-task")

				_, err := permClient.CreateRole(context.Background(), roleName, permission)
				Expect(err).NotTo(HaveOccurred())

				defer permClient.DeleteRole(context.Background(), roleName)

				err = permClient.AssignRole(context.Background(), roleName, actor)
				Expect(err).NotTo(HaveOccurred())

				tasks, err := cc.V3ListTasks(user.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(tasks).To(HaveLen(0))
			})
		})
	})
})
