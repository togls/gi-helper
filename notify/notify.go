package notify

import (
	"context"

	"github.com/togls/gi-helper/check"
)

type Notifier interface {
	Notify(ctx context.Context, msg check.Message) error
}
