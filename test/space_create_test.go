package test_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#SpaceCreate", func() {
	It("should be able to GET /v2/info", func() {
		var v2InfoResp struct {
			Name string `json:"name"`
		}

		res, err := cc.Get("/v2/info", "", &v2InfoResp)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.StatusCode).To(Equal(200))
		Expect(v2InfoResp.Name).To(Equal("vcap"))
	})

	It("should be able to GET from a restricted endpoint", func() {
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

		token, err := createSignedToken()
		Expect(err).NotTo(HaveOccurred())

		res, err := cc.Get("/v2/organizations", token.AccessToken, &listOrgsResp)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.StatusCode).To(Equal(200))
		Expect(listOrgsResp.TotalResults).To(Equal(2))
	})

	XIt("should be able to GET from a restricted endpoint", func() {
		body := struct {
			Name string `json:"name"`
		}{
			Name: "org-foo",
		}

		bodyBits, err := json.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("%s\n", bodyBits)

		var createOrgResp struct {
			Metadata struct {
				GUID string `json:"guid"`
			} `json:"metadata"`
			Entity struct {
				Name string `json:"name"`
			} `json:"entity"`
		}

		token, err := createSignedToken()
		Expect(err).NotTo(HaveOccurred())

		res, err := cc.Post("/v2/organizations", token.AccessToken, bodyBits, &createOrgResp)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.StatusCode).To(Equal(200))
		Expect(createOrgResp.Entity.Name).To(Equal("foo"))
	})
})
