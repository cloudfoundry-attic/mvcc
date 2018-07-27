package test_test

import (
	"fmt"

	"code.cloudfoundry.org/mvcc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Baselines", func() {
	Describe("baseline API behavior", func() {
		Context("when the user is an OrgManager", func() {
			var (
				org mvcc.Organization
			)

			BeforeEach(func() {
				var err error
				org, err = cc.V3CreateOrganization(admin.AccessToken)
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
})
