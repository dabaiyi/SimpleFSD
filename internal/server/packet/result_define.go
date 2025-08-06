package packet

type Result struct {
	success bool
	errno   ClientError
	fatal   bool
	env     string
}

func ResultSuccess() *Result {
	return &Result{
		success: true,
		errno:   Ok,
		fatal:   false,
		env:     "",
	}
}

func ResultError(errno ClientError, fatal bool, env string) *Result {
	return &Result{
		success: false,
		errno:   errno,
		fatal:   fatal,
		env:     env,
	}
}
