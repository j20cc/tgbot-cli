package polling

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/example/tgbot-cli/internal/telegram"
)

type telegramAPI interface {
	DeleteWebhook(ctx context.Context) error
	GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]telegram.Update, error)
}

type Options struct {
	Interval      time.Duration
	TimeoutSecond int
	InitialOffset int64
	DeleteWebhook bool
	Once          bool
}

type Poller struct {
	api  telegramAPI
	opts Options
}

func New(api telegramAPI, opts Options) *Poller {
	if opts.Interval < 0 {
		opts.Interval = 0
	}
	if opts.TimeoutSecond < 0 {
		opts.TimeoutSecond = 0
	}
	return &Poller{api: api, opts: opts}
}

func (p *Poller) Run(ctx context.Context, outWriter, errWriter io.Writer) error {
	offset := p.opts.InitialOffset
	if p.opts.DeleteWebhook {
		if _, err := fmt.Fprintln(errWriter, "[info] deleting webhook before polling..."); err != nil {
			return err
		}
		if err := p.api.DeleteWebhook(ctx); err != nil {
			return err
		}
	}

	for {
		updates, err := p.api.GetUpdates(ctx, offset, p.opts.TimeoutSecond)
		if err != nil {
			return err
		}

		for _, update := range updates {
			if _, err := outWriter.Write(append(update.Raw, '\n')); err != nil {
				return err
			}
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
		}

		if p.opts.Once {
			return nil
		}

		if p.opts.Interval > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.opts.Interval):
			}
		}
	}
}
