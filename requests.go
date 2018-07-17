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
