package testkit

import (
	"context"
	"time"

	"github.com/tochka/core-kit/ping"
)

func Wait(pinger ping.Pinger, timeout time.Duration) error {
	const count = 10
	sleep := timeout / time.Duration(count)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var err error
	for {
		select {
		case <-timer.C:
			return err
		default:
			err = pinger.Ping(context.Background())
			if err == nil {
				return nil
			}
			time.Sleep(sleep)
		}
	}
}
