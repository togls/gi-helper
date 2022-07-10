package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/togls/gi-helper/check"
	cc "github.com/togls/gi-helper/context"
)

const (
	tgBaseApi = "https://api.telegram.org/bot%s/sendMessage"
)

type tg struct {
	token  string
	userID string

	client *http.Client
}

func NewTGNotifier(token, userID string) Notifier {
	if token == "" {
		panic("token is empty")
	}

	if userID == "" {
		panic("userID is empty")
	}

	return &tg{
		token:  token,
		userID: userID,
	}
}

func (tg *tg) Notify(ctx context.Context, msg check.Message) error {
	if tg.client == nil {
		tg.client = cc.Client(ctx)
	}

	url := fmt.Sprintf(tgBaseApi, tg.token)

	data := map[string]string{
		"chat_id": tg.userID,
		"text":    fmt.Sprintf("%s\n\n%s", msg.Title(), msg.Content()),
	}

	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := tg.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code is %d", resp.StatusCode)
	}

	return nil
}
