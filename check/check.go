package check

import (
	"context"
	"time"
)

type Message interface {
	Title() string
	Content() string
}

type Checker interface {
	Check(ctx context.Context) (msg Message, err error)
}

type SpecifiedChecker interface {
	Checker
	Next() <-chan time.Time
}
