package worker

import "github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/plug"

type Worker interface {
	plug.InteractUser
	plug.CronUser
}
