package plug

type ConfigChecker interface {
	CheckConfig() error
}

type Stopper interface {
	Stop()
}

type Statuser interface {
	GetStatus() bool
}
