package packet

import "errors"

type Result struct {
	success bool
	errno   ClientError
	fatal   bool
	env     string
	err     error
}

var defaultError = errors.New("no details error provided")

func resultSuccess() *Result {
	return &Result{
		success: true,
		errno:   Ok,
		fatal:   false,
		env:     "",
		err:     nil,
	}
}

func resultError(errno ClientError, fatal bool, env string, err error) *Result {
	result := &Result{
		success: false,
		errno:   errno,
		fatal:   fatal,
		env:     env,
	}
	if err != nil {
		result.err = err
	} else {
		result.err = defaultError
	}
	return result
}
