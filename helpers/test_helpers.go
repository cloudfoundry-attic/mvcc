package helpers

import "fmt"

func OrgResourceID(orgUUID string) string {
	return fmt.Sprintf("%s/*", orgUUID)
}

func SpaceResourceID(orgUUID, spaceUUID string) string {
	return fmt.Sprintf("%s/%s", orgUUID, spaceUUID)
}
