package common

import "errors"

var (
	ErrUnknownTask     = errors.New("unknown task")
	ErrUnknownWorkflow = errors.New("unknown workflow")
)
