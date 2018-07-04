package mvcc

type V2OrganizationResponse struct {
	Metadata struct {
		GUID string `json:"guid"`
	} `json:"metadata"`
	Entity struct {
		Status string `json:"status"`
	} `json:"entity"`
}

type V2SpaceResponse struct {
	Metadata struct {
		GUID string `json:"guid"`
	} `json:"metadata"`
	Entity struct {
		Name string `json:"name"`
	} `json:"entity"`
}
