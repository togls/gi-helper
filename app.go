package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/togls/gi-helper/check"
	"github.com/togls/gi-helper/notify"
)

type Application struct {
	Config *Config

	cs []check.Checker
	ns []notify.Notifier
}

func (app *Application) run(ctx context.Context) error {
	checkTime := os.Getenv("CHECK_TIME")
	if checkTime == "" {
		checkTime = "06:00"
	}

	t := time.Now()
	ctime, err := time.Parse("15:04", checkTime)
	if err != nil {
		return err
	}

	n := time.Date(t.Year(), t.Month(), t.Day(),
		ctime.Hour(), ctime.Minute(), 0, 0,
		t.Location())

	d := n.Sub(t)
	if d < 0 {
		n = n.Add(24 * time.Hour)
		d = n.Sub(t)
	}

	tick := time.After(d)
	once := time.After(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-once: // run immediately
		case <-tick:
			tick = time.After(24 * time.Hour)
		}

		for _, c := range app.cs {
			msg, err := c.Check(ctx)
			if err != nil {
				fmt.Printf("check error: %s\n", err)
				continue
			}

			for _, n := range app.ns {
				err := n.Notify(ctx, msg)
				if err != nil {
					fmt.Printf("notify error: %s\n", err)
					continue
				}
			}
		}
	}
}
