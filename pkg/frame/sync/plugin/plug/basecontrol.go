package plug

type ConfigChecker interface {
	CheckConfig() error
}

type Stopper interface {
	Stop()
}

type SuccessCounter interface {
	AddSuccess()
}

type ErrorCounter interface {
	AddError()
}
