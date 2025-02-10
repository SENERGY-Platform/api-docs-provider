package util

import (
	"context"
	"os"
	"os/signal"
)

func WaitForSignal(ctx context.Context, signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	for _, sig := range signals {
		signal.Notify(ch, sig)
	}
	select {
	case sig := <-ch:
		Logger.Warningf("caught signal '%s'", sig)
		break
	case <-ctx.Done():
		break
	}
	signal.Stop(ch)
}
