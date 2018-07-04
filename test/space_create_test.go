package test_test

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/mvcc"
	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#SpaceCreate", func() {
	Describe("baseline API behavior", func() {
		Context("when the user is an OrgManager", func() {
			var (
				org mvcc.Organization
			)

			BeforeEach(func() {
				var err error
				org, err = cc.CreateRandomOrganization(admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())

				associateOrgManagerPath := fmt.Sprintf("/v2/organizations/%s/managers/%s", org.UUID, user.UUID)
				_, err = cc.Put(associateOrgManagerPath, admin.AccessToken, struct{}{}, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				orgURL := fmt.Sprintf("/v2/organizations/%s?recursive=true", org.UUID)
				res, err := cc.Delete(orgURL, admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(204))
			})

			It("can create a space", func() {
				spaceName := randomName("space")
				body := mvcc.V2SpaceRequest{
					Name:             spaceName,
					OrganizationGUID: org.UUID,
				}

				var createSpaceResp mvcc.V2SpaceResponse
				res, err := cc.Post("/v2/spaces", user.AccessToken, body, &createSpaceResp)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.StatusCode).To(Equal(201))
				Expect(createSpaceResp.Entity.Name).To(Equal(spaceName))
			})

			Context("when the organization is suspended", func() {
				BeforeEach(func() {
					orgPath := fmt.Sprintf("/v2/organizations/%s", org.UUID)

					var orgResp mvcc.V2OrganizationResponse

					res, err := cc.Get(orgPath, admin.AccessToken, &orgResp)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(200))
					Expect(orgResp.Entity.Status).To(Equal("active"))

					body := mvcc.V2OrganizationRequest{
						Status: "suspended",
					}

					res, err = cc.Put(orgPath, admin.AccessToken, body, &orgResp)
					Expect(err).ToNot(HaveOccurred())

					Expect(res.StatusCode).To(Equal(201))
					Expect(orgResp.Entity.Status).To(Equal("suspended"))
				})

				It("they can NOT create a space", func() {
					spaceName := randomName("space")
					body := mvcc.V2SpaceRequest{
						Name:             spaceName,
						OrganizationGUID: org.UUID,
					}

					var createSpaceResp mvcc.V2SpaceResponse

					res, err := cc.Post("/v2/spaces", user.AccessToken, body, &createSpaceResp)
					Expect(err).NotTo(HaveOccurred())

					Expect(res.StatusCode).To(Equal(403))
				})
			})
		})
	})

	Describe("FGP space.create", func() {
		Context("when the user has the space.create permission in an org", func() {
			var (
				roleName string
				org      mvcc.Organization
			)

			BeforeEach(func() {
				var err error
				org, err = cc.CreateRandomOrganization(admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())

				permission := perm.Permission{
					Action:          "space.create",
					ResourcePattern: org.UUID,
				}

				roleName = randomName("role")
				_, err = permClient.CreateRole(context.Background(), roleName, permission)
				Expect(err).NotTo(HaveOccurred())

				err = permClient.AssignRole(context.Background(), roleName, actor)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				orgURL := fmt.Sprintf("/v2/organizations/%s?recursive=true", org.UUID)
				res, err := cc.Delete(orgURL, admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(204))

				permClient.DeleteRole(context.Background(), roleName)
			})

			It("they can create a space", func() {
				body := mvcc.V2SpaceRequest{
					Name:             "my-space",
					OrganizationGUID: org.UUID,
				}
				res, err := cc.Post("/v2/spaces", user.AccessToken, body, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(201))
			})

			Context("when the org is suspended", func() {
				BeforeEach(func() {
					body := mvcc.V2OrganizationRequest{
						Status: "suspended",
					}

					var orgUpdateResponse mvcc.V2OrganizationResponse

					orgURL := fmt.Sprintf("/v2/organizations/%s", org.UUID)
					res, err := cc.Put(orgURL, admin.AccessToken, body, &orgUpdateResponse)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(201))
					Expect(orgUpdateResponse.Entity.Status).To(Equal("suspended"))
				})

				It("they can NOT create a space", func() {
					body := mvcc.V2SpaceRequest{
						Name:             "my-space",
						OrganizationGUID: org.UUID,
					}
					res, err := cc.Post("/v2/spaces", user.AccessToken, body, nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(403))
				})
			})
		})

		Context("when the user is authorized and not an Org Manger, admin, or space.create-er", func() {
			var (
				org mvcc.Organization
			)

			BeforeEach(func() {
				var err error
				org, err = cc.CreateRandomOrganization(admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				orgURL := fmt.Sprintf("/v2/organizations/%s?recursive=true", org.UUID)
				res, err := cc.Delete(orgURL, admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(204))
			})

			It("they can NOT create a space", func() {
				body := mvcc.V2SpaceRequest{
					Name:             "my-space",
					OrganizationGUID: org.UUID,
				}
				res, err := cc.Post("/v2/spaces", user.AccessToken, body, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(403))
			})

			Context("when the org is suspended", func() {
				BeforeEach(func() {
					body := mvcc.V2OrganizationRequest{
						Status: "suspended",
					}

					var orgUpdateResponse mvcc.V2OrganizationResponse

					orgURL := fmt.Sprintf("/v2/organizations/%s", org.UUID)
					res, err := cc.Put(orgURL, admin.AccessToken, body, &orgUpdateResponse)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(201))
					Expect(orgUpdateResponse.Entity.Status).To(Equal("suspended"))
				})

				It("they can NOT create a space", func() {
					body := mvcc.V2SpaceRequest{
						Name:             "my-space",
						OrganizationGUID: org.UUID,
					}
					res, err := cc.Post("/v2/spaces", user.AccessToken, body, nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(403))
				})
			})
		})
	})
})
