package mvcc

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

type PackageType string

const (
	BitsType   PackageType = "bits"
	DockerType PackageType = "docker"
)

type User struct {
	UUID        string
	AccessToken string
}

type Organization struct {
	Name string
	UUID string
}

type Space struct {
	Name string
	UUID string
}

type App struct {
	Name string
	UUID string
}

type Package struct {
	UUID  string
	Type  PackageType
	State string
}

type Build struct {
	UUID        string
	State       string
	DropletUUID string
}

type Task struct {
	UUID string
}

func RandomUUID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.NewV4().String())
}
