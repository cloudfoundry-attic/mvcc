package mvcc

type V3Error struct {
	Code   int    `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type V3ErrorResponse struct {
	Errors []V3Error `json:"errors"`
}

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

type v3PackageResponse struct {
	GUID  string      `json:"guid"`
	Type  PackageType `json:"type"`
	State string      `json:"state"`
}

type v3BuildResponse struct {
	GUID    string `json:"guid"`
	State   string `json:"state"`
	Droplet struct {
		GUID string `json:"guid"`
	} `json:"droplet"`
}

type v3TaskResponse struct {
	Name        string `json:"name"`
	GUID        string `json:"guid"`
	DropletGUID string `json:"droplet_guid"`
}

type v3ListTasksResponse struct {
	Resources []v3TaskResponse `json:"resources"`
}
