package mvcc

import (
	"errors"
	"fmt"
)

var (
	ErrFailedToStart              = errors.New("failed to start")
	ErrCCBinaryPathNotSet         = errors.New("the filepath to the CC binary must be set")
	ErrCCConfigPathNotSet         = errors.New("the filepath to the CC config file must be set")
	ErrPermServerBinaryPathNotSet = errors.New("the filepath to the perm binary must be set")
	ErrPermServerCertsPathNotSet  = errors.New("the filepath to the perm TLS certificates must be set")

	ErrBadRequest          = errors.New("bad request")
	ErrUnauthenticated     = errors.New("unauthenticated")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrUnprocessableEntity = errors.New("unprocessable entity")

	ErrInternalServer = errors.New("internal server error")
	ErrBadGateway     = errors.New("bad gateway")
)

type ErrUnexpectedStatusCode struct {
	StatusCode int
}

func (e *ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("unexpected error code (%d)", e.StatusCode)
}
