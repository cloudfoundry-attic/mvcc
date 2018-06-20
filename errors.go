package mvcc

import "errors"

var (
	ErrFailedToStart      = errors.New("failed to start")
	ErrCCBinaryPathNotSet = errors.New("the filepath to the CC binary must be set")
	ErrCCConfigPathNotSet = errors.New("the filepath to the CC config file must be set")
)
