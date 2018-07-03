package test_test

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/perm/pkg/perm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#SpaceCreate", func() {
	Describe("baseline API behavior", func() {

		Context("when the user is an admin", func() {
			It("can GET from an unprotected endpoint", func() {
				var v2InfoResp struct {
					Name string `json:"name"`
				}

				res, err := cc.Get("/v2/info", "", &v2InfoResp)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.StatusCode).To(Equal(200))
				Expect(v2InfoResp.Name).To(Equal("vcap"))
			})

			It("can GET from a protected endpoint", func() {
				var listOrgsResp struct {
					TotalResults int `json:"total_results"`
					Resources    []struct {
						Metadata struct {
							GUID string `json:"guid"`
						} `json:"metadata"`
						Entity struct {
							Name string `json:"name"`
						} `json:"entity"`
					} `json:"resources"`
				}

				res, err := cc.Get("/v2/organizations", admin.AccessToken, &listOrgsResp)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.StatusCode).To(Equal(200))
			})

			It("can POST to a protected endpoint", func() {
				orgName := randomName("org")
				body := struct {
					Name string `json:"name"`
				}{
					Name: orgName,
				}

				var createOrgResp struct {
					Metadata struct {
						GUID string `json:"guid"`
					} `json:"metadata"`
					Entity struct {
						Name string `json:"name"`
					} `json:"entity"`
				}

				res, err := cc.Post("/v2/organizations", admin.AccessToken, body, &createOrgResp)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.StatusCode).To(Equal(201))
				Expect(createOrgResp.Entity.Name).To(Equal(orgName))
				orgGuid := createOrgResp.Metadata.GUID

				orgURL := fmt.Sprintf("/v2/organizations/%s?recursive=true", orgGuid)
				res, err = cc.Delete(orgURL, admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(204))
			})
		})

		Context("when the user is an OrgManager", func() {
			var (
				orgGuid string
				err     error
			)

			BeforeEach(func() {
				orgName := randomName("org")
				orgCreateBody := struct {
					Name string `json:"name"`
				}{
					Name: orgName,
				}

				var createOrgResp struct {
					Metadata struct {
						GUID string `json:"guid"`
					} `json:"metadata"`
					Entity struct {
						Name string `json:"name"`
					} `json:"entity"`
				}

				_, err = cc.Post("/v2/organizations", admin.AccessToken, orgCreateBody, &createOrgResp)
				Expect(err).NotTo(HaveOccurred())

				orgGuid = createOrgResp.Metadata.GUID

				associateOrgManagerPath := fmt.Sprintf("/v2/organizations/%s/managers/%s", orgGuid, user.UUID)
				_, err = cc.Put(associateOrgManagerPath, admin.AccessToken, struct{}{}, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				orgURL := fmt.Sprintf("/v2/organizations/%s?recursive=true", orgGuid)
				res, err := cc.Delete(orgURL, admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(204))
			})

			It("can create a space", func() {
				spaceName := randomName("space")
				body := struct {
					Name             string `json:"name"`
					OrganizationGUID string `json:"organization_guid"`
				}{
					Name:             spaceName,
					OrganizationGUID: orgGuid,
				}

				var createSpaceResp struct {
					Metadata struct {
						GUID string `json:"guid"`
					} `json:"metadata"`
					Entity struct {
						Name string `json:"name"`
					} `json:"entity"`
				}

				res, err := cc.Post("/v2/spaces", user.AccessToken, body, &createSpaceResp)
				Expect(err).NotTo(HaveOccurred())

				Expect(res.StatusCode).To(Equal(201))
				Expect(createSpaceResp.Entity.Name).To(Equal(spaceName))
			})

			Context("when the organization is suspended", func() {
				BeforeEach(func() {
					orgPath := fmt.Sprintf("/v2/organizations/%s", orgGuid)

					var orgResp struct {
						Entity struct {
							Status string `json:"status"`
						} `json:"entity"`
					}

					res, err := cc.Get(orgPath, admin.AccessToken, &orgResp)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(200))
					Expect(orgResp.Entity.Status).To(Equal("active"))

					body := struct {
						Status string `json:"status"`
					}{
						Status: "suspended",
					}

					res, err = cc.Put(orgPath, admin.AccessToken, body, &orgResp)
					Expect(err).ToNot(HaveOccurred())

					Expect(res.StatusCode).To(Equal(201))
					Expect(orgResp.Entity.Status).To(Equal("suspended"))
				})

				It("they can NOT create a space", func() {
					spaceName := randomName("space")
					body := struct {
						Name             string `json:"name"`
						OrganizationGUID string `json:"organization_guid"`
					}{
						Name:             spaceName,
						OrganizationGUID: orgGuid,
					}

					var createSpaceResp struct {
						Metadata struct {
							GUID string `json:"guid"`
						} `json:"metadata"`
						Entity struct {
							Name string `json:"name"`
						} `json:"entity"`
					}

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
				orgGuid  string
				roleName string
			)

			BeforeEach(func() {
				orgName := randomName("org")
				body := struct {
					Name string `json:"name"`
				}{
					Name: orgName,
				}

				var orgCreateResponse struct {
					Metadata struct {
						GUID string `json:"guid"`
					} `json:"metadata"`
				}

				res, err := cc.Post("/v2/organizations", admin.AccessToken, body, &orgCreateResponse)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(201))

				orgGuid = orgCreateResponse.Metadata.GUID
				permission := perm.Permission{
					Action:          "space.create",
					ResourcePattern: orgGuid,
				}

				roleName = randomName("role")
				_, err = permClient.CreateRole(context.Background(), roleName, permission)
				Expect(err).NotTo(HaveOccurred())

				actor := perm.Actor{
					ID:        user.UUID,
					Namespace: validIssuer,
				}

				err = permClient.AssignRole(context.Background(), roleName, actor)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				orgURL := fmt.Sprintf("/v2/organizations/%s?recursive=true", orgGuid)
				res, err := cc.Delete(orgURL, admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(204))

				permClient.DeleteRole(context.Background(), roleName)
			})

			It("they can create a space", func() {
				body := struct {
					Name    string `json:"name"`
					OrgGuid string `json:"organization_guid"`
				}{
					Name:    "my-space",
					OrgGuid: orgGuid,
				}
				res, err := cc.Post("/v2/spaces", user.AccessToken, body, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(201))
			})

			Context("when the org is suspended", func() {
				BeforeEach(func() {
					body := struct {
						Status string `json:"status"`
					}{
						Status: "suspended",
					}

					var orgUpdateResponse struct {
						Metadata struct {
							GUID string `json:"guid"`
						} `json:"metadata"`
						Entity struct {
							Status string `json:"status"`
						} `json:"entity"`
					}

					orgURL := fmt.Sprintf("/v2/organizations/%s", orgGuid)
					res, err := cc.Put(orgURL, admin.AccessToken, body, &orgUpdateResponse)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(201))
					Expect(orgUpdateResponse.Entity.Status).To(Equal("suspended"))
				})

				It("they can NOT create a space", func() {
					body := struct {
						Name    string `json:"name"`
						OrgGuid string `json:"organization_guid"`
					}{
						Name:    "my-space",
						OrgGuid: orgGuid,
					}
					res, err := cc.Post("/v2/spaces", user.AccessToken, body, nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(403))
				})
			})
		})

		Context("when the user is authorized and not an Org Manger, admin, or space.create-er", func() {
			var (
				orgGuid string
			)
			BeforeEach(func() {
				orgName := randomName("org")
				body := struct {
					Name string `json:"name"`
				}{
					Name: orgName,
				}

				var orgCreateResponse struct {
					Metadata struct {
						GUID string `json:"guid"`
					} `json:"metadata"`
				}

				res, err := cc.Post("/v2/organizations", admin.AccessToken, body, &orgCreateResponse)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(201))

				orgGuid = orgCreateResponse.Metadata.GUID
			})

			AfterEach(func() {
				orgURL := fmt.Sprintf("/v2/organizations/%s?recursive=true", orgGuid)
				res, err := cc.Delete(orgURL, admin.AccessToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(204))
			})

			It("they can NOT create a space", func() {
				body := struct {
					Name    string `json:"name"`
					OrgGuid string `json:"organization_guid"`
				}{
					Name:    "my-space",
					OrgGuid: orgGuid,
				}
				res, err := cc.Post("/v2/spaces", user.AccessToken, body, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(403))
			})

			Context("when the org is suspended", func() {
				BeforeEach(func() {
					body := struct {
						Status string `json:"status"`
					}{
						Status: "suspended",
					}

					var orgUpdateResponse struct {
						Metadata struct {
							GUID string `json:"guid"`
						} `json:"metadata"`
						Entity struct {
							Status string `json:"status"`
						} `json:"entity"`
					}

					orgURL := fmt.Sprintf("/v2/organizations/%s", orgGuid)
					res, err := cc.Put(orgURL, admin.AccessToken, body, &orgUpdateResponse)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(201))
					Expect(orgUpdateResponse.Entity.Status).To(Equal("suspended"))
				})

				It("they can NOT create a space", func() {
					body := struct {
						Name    string `json:"name"`
						OrgGuid string `json:"organization_guid"`
					}{
						Name:    "my-space",
						OrgGuid: orgGuid,
					}
					res, err := cc.Post("/v2/spaces", user.AccessToken, body, nil)
					Expect(err).NotTo(HaveOccurred())
					Expect(res.StatusCode).To(Equal(403))
				})
			})
		})
	})
})
