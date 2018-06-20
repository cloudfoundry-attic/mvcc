package test_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#SpaceCreate", func() {
	It("should be able to GET /v2/info", func() {
		var v2InfoResp struct {
			Name string `json:"name"`
		}

		res, err := cc.Get("/v2/info", &v2InfoResp)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.StatusCode).To(Equal(200))
		Expect(v2InfoResp.Name).To(Equal("vcap"))
	})

	XIt("should be able to create a space", func() {
		createOrgBody := struct {
			Name string `json:"name"`
		}{
			Name: "foo",
		}

		var createOrgResp struct {
			Metadata struct {
				GUID string `json:"guid"`
			} `json:"metadata"`
			Entity struct {
				Name string `json:"name"`
			} `json:"entity"`
		}

		res, err := cc.Post("/v2/organizations", createOrgBody, &createOrgResp)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.StatusCode).To(Equal(200))
		Expect(createOrgResp.Entity.Name).To(Equal("foo"))
	})
})
