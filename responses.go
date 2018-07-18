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

type v3OrganizationResponse struct {
	Name string `json:"name"`
	GUID string `json:"guid"`
}

type v3SpaceResponse struct {
	Name string `json:"name"`
	GUID string `json:"guid"`
}

type v3AppResponse struct {
	Name string `json:"name"`
	GUID string `json:"guid"`
}

type v3TaskResponse struct {
	Name        string `json:"name"`
	GUID        string `json:"guid"`
	DropletGUID string `json:"droplet_guid"`
}
