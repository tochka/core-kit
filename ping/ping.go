package ping

import (
	"context"
	"sync"
)

type Pinger interface {
	Ping(ctx context.Context) error
}
type PingerFunc func(ctx context.Context) error

func (f PingerFunc) Ping(ctx context.Context) error {
	return f(ctx)
}

func NewOncePinger(p Pinger) *OncePinger {
	return &OncePinger{
		once:   false,
		mtx:    sync.Mutex{},
		pinger: p,
	}
}

type OncePinger struct {
	once   bool
	mtx    sync.Mutex
	pinger Pinger
}

func (p *OncePinger) Ping(ctx context.Context) error {
	if p.once == false {
		p.mtx.Lock()
		defer p.mtx.Unlock()

		if p.once == false {
			if err := p.pinger.Ping(ctx); err != nil {
				return err
			}
			p.once = true
		}
	}
	return nil
}
