package check

import (
	"context"
)

type Message interface {
	Title() string
	Content() string
}

type Checker interface {
	Check(ctx context.Context) (msg Message, err error)
}
