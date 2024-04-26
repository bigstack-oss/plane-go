package transit

import (
	"context"

	"sync"

	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/plug"
)

var (
	Plugin = make(map[string]Transit)
)

type Transit interface {
	plug.Parter
	DoTransit()
}

func WrapWithSingleMsgLoop(ctx context.Context, wg *sync.WaitGroup, group string, coreFunc func([]byte) (protocol.Job, error)) func() {
	return func() {
		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, isChnOpen := <-plugin.I2TChan[group]:
				switch isChnOpen {
				case true:
					task, err := coreFunc(msg)
					if err != nil {
						continue
					}

					plugin.T2PChan[group] <- task
				case false:
					close(plugin.T2PChan[group])
					return
				}
			}
		}
	}
}

func WrapWithBatchMsgLoop(ctx context.Context, wg *sync.WaitGroup, group string, coreFunc func([]byte) ([]protocol.Job, error)) func() {
	return func() {
		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, isChnOpen := <-plugin.I2TChan[group]:
				switch isChnOpen {
				case true:
					tasks, err := coreFunc(msg)
					if err != nil || len(tasks) == 0 {
						continue
					}

					for _, task := range tasks {
						plugin.T2PChan[group] <- task
					}
				case false:
					close(plugin.T2PChan[group])
					return
				}
			}
		}
	}
}
