package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	cc "github.com/togls/gi-helper/context"
)

func main() {
	configFile := flag.String("c", "config.json", "config file")
	flag.Parse()

	if _, err := os.Stat(*configFile); err != nil {
		fmt.Printf("config file not found: %s\n", *configFile)
		os.Exit(1)
	}

	app, err := parseConfig(*configFile)
	if err != nil {
		fmt.Printf("parse config error: %s\n", err)
		os.Exit(1)
	}

	ctx, cancle := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancle()

	ctx = cc.WithClient(ctx, &http.Client{})

	if err := app.run(ctx); err != nil {
		fmt.Printf("run error: %s\n", err)
		os.Exit(1)
	}

	<-ctx.Done()
}
