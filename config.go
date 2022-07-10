package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/togls/gi-helper/check/ys"
	"github.com/togls/gi-helper/notify"
)

type Config struct {
	BBS    []string            `json:"bbs"`
	Cloud  []map[string]string `json:"cloud"`
	Notify []struct {
		Type   string            `json:"type"`
		Params map[string]string `json:"params"`
	} `json:"notify"`
}

func parseConfig(path string) (*Application, error) {
	var config Config
	if err := parseJSON(path, &config); err != nil {
		return nil, err
	}

	app := &Application{
		Config: &config,
	}

	for _, cookie := range config.BBS {
		app.cs = append(app.cs, ys.New(cookie))
	}

	for _, c := range config.Cloud {
		app.cs = append(app.cs, ys.NewCloud(c))
	}

	for _, n := range config.Notify {
		switch n.Type {
		case "telegram":
			token := n.Params["token"]
			userID := n.Params["userid"]

			app.ns = append(app.ns, notify.NewTGNotifier(token, userID))
		default:
			fmt.Printf("unknown notify type: %s\n", n.Type)
		}
	}

	return app, nil
}

func parseJSON(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	return d.Decode(v)
}
