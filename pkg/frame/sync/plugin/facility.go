package plugin

const (
	Interact     = "interact"
	StageTransit = "transit"
	StageProcess = "process"
	StageRequest = "request"

	Stage = "stage"

	CronJob = "cronjobs"
)

var (
	Service string
	Metrics = &Metric{}

	ActivatedTransit []string
	ActivatedProcess []string
	ActivatedRequest []string
)

type Metric struct {
	InteractOK  uint64 `json:"interactOK"`
	InteractErr uint64 `json:"interactErr"`

	TransitOK  uint64 `json:"transitOK"`
	TransitErr uint64 `json:"transitErr"`

	ProcessOK  uint64 `json:"processOK"`
	ProcessErr uint64 `json:"processErr"`

	RequestOK  uint64 `json:"requestOK"`
	RequestErr uint64 `json:"requestErr"`
}
