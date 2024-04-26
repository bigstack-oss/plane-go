package plugin

import "github.com/bigstack-oss/plane-go/pkg/base/protocol"

const (
	Input   = "input"
	Transit = "transit"
	Process = "process"
	Output  = "output"
	CronJob = "cronjobs"
)

var (
	Service string

	ChanSize int32
	I2TChan  = make(map[string](chan []byte))
	T2PChan  = make(map[string](chan protocol.Job))
	P2OChan  = make(map[string](chan protocol.Job))

	Metrics = &Metric{}
	Records = &Record{}

	InputDone   *int64 = new(int64)
	TransitDone *int64 = new(int64)
	ProcessDone *int64 = new(int64)
	OutputDone  *int64 = new(int64)

	IsOneTimeExec = false
)

type Metric struct {
	InputOK  int64 `json:"inputOK"`
	InputErr int64 `json:"inputErr"`

	TransitOK  int64 `json:"transitOK"`
	TransitErr int64 `json:"transitErr"`

	ProcessOK  int64 `json:"processOK"`
	ProcessErr int64 `json:"processErr"`

	OutputOK  int64 `json:"outputOK"`
	OutputErr int64 `json:"outputErr"`
}

type Record struct {
	Input   []protocol.Job
	Transit []protocol.Job
	Process []protocol.Job
	Output  []protocol.Job
}

type inputStatus struct {
	Completed bool
}

type transitStatus struct {
	Completed bool
}
type processStatus struct {
	Completed bool
}

type outputStatus struct {
	Completed bool
}
