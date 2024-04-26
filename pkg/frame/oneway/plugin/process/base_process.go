package process

import (
	"context"
	"sync"

	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/plug"
)

var (
	Plugin = make(map[string]Process)
)

type Process interface {
	plug.Parter
	DoProcess()
}

func WrapWithSingleMsgLoop(ctx context.Context, wg *sync.WaitGroup, group string, coreFunc func(protocol.Job, bool) (protocol.Job, error)) func() {
	return func() {
		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, isChnOpen := <-plugin.T2PChan[group]:
				switch isChnOpen {
				case true:
					msg, err := coreFunc(msg, isChnOpen)
					if err != nil {
						continue
					}

					plugin.P2OChan[group] <- msg
				case false:
					close(plugin.P2OChan[group])
					return
				}
			}
		}
	}
}

func WrapWithBatchMsgLoop(ctx context.Context, wg *sync.WaitGroup, group string, coreFunc func(protocol.Job, bool) ([]protocol.Job, error)) func() {
	return func() {
		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, isChnOpen := <-plugin.T2PChan[group]:
				switch isChnOpen {
				case true:
					msgs, err := coreFunc(msg, isChnOpen)
					if err != nil {
						continue
					}

					for _, msg := range msgs {
						plugin.P2OChan[group] <- msg
					}
				case false:
					close(plugin.P2OChan[group])
					return
				}
			}
		}
	}
}
