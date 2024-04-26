package input

import (
	"context"
	"sync"
	"time"

	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/plug"
)

var (
	Plugin = make(map[string]Input)
)

type Input interface {
	plug.Parter
	DoInput()
}

func WrapWithSingleMsgLoop(ctx context.Context, wg *sync.WaitGroup, group string, coreFunc func() ([]byte, error), interval time.Duration) func() {
	return func() {
		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := coreFunc()
				if err != nil {
					continue
				}

				plugin.I2TChan[group] <- msg

				if plugin.IsOneTimeExec {
					close(plugin.I2TChan[group])
					return
				}

				time.Sleep(interval * time.Second)
			}
		}
	}
}

func WrapWithBatchMsgLoop(ctx context.Context, wg *sync.WaitGroup, group string, coreFunc func() ([][]byte, error), interval time.Duration) func() {
	return func() {
		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msgs, err := coreFunc()
				if err != nil {
					continue
				}

				for _, msg := range msgs {
					plugin.I2TChan[group] <- msg
				}

				if plugin.IsOneTimeExec {
					close(plugin.I2TChan[group])
					return
				}

				time.Sleep(interval * time.Second)
			}
		}
	}
}
