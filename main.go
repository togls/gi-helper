package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"

	cc "github.com/togls/gi-helper/context"
	"github.com/togls/gi-helper/log"
)

func main() {
	configFile := flag.String("c", "config.json", "config file")
	flag.Parse()

	if _, err := os.Stat(*configFile); err != nil {
		log.Fatal().Err(err).Msg("config file not found")
	}

	app, err := parseConfig(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("parse config error")
	}

	ctx, cancle := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancle()

	ctx = cc.WithClient(ctx, &http.Client{})

	if err := app.run(ctx); err != nil {
		log.Fatal().Err(err).Msg("app run")
	}

	<-ctx.Done()
}
