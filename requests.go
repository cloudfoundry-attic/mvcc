package mvcc

type V2OrganizationRequest struct {
	Status string `json:"status"`
}

type V2SpaceRequest struct {
	Name             string `json:"name"`
	OrganizationGUID string `json:"organization_guid"`
}

type v3OrganizationRequest struct {
	Name string `json:"name"`
}

type v3SpaceRequest struct {
	Name          string `json:"name"`
	Relationships struct {
		Organization struct {
			Data struct {
				GUID string `json:"guid"`
			} `json:"data"`
		} `json:"organization"`
	} `json:"relationships"`
}

type v3AppRequest struct {
	Name      string `json:"name"`
	Lifecycle struct {
		Type string   `json:"type"`
		Data struct{} `json:"data"`
	} `json:"lifecycle"`
	Relationships struct {
		Space struct {
			Data struct {
				GUID string `json:"guid"`
			} `json:"data"`
		} `json:"space"`
	} `json:"relationships"`
}

type v3PackageRequest struct {
	Type          PackageType `json:"type"`
	Relationships struct {
		App struct {
			Data struct {
				GUID string `json:"guid"`
			} `json:"data"`
		} `json:"app"`
	} `json:"relationships"`
	Data struct {
		Image string `json:"image"`
	} `json:"data"`
}

type v3BuildRequest struct {
	Package struct {
		GUID string `json:"guid"`
	} `json:"package"`
}

type v3TaskRequest struct {
	Command     string `json:"command"`
	DropletGUID string `json:"droplet_guid"`
}

type v2FeatureFlagRequest struct {
	Enabled bool `json:"enabled"`
}
