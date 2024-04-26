package worker

import "github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/plug"

type Worker interface {
	plug.PartUser
	plug.CronUser
	plug.Statuser
}
