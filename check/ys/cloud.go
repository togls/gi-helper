package ys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/togls/gi-helper/check"
	cc "github.com/togls/gi-helper/context"
)

const (
	cloudApi = "https://api-cloudgame.mihoyo.com/hk4e_cg_cn/wallet/wallet/get"
)

type Cloud struct {
	headers http.Header

	client *http.Client
}

func NewCloud(headers map[string]string) *Cloud {

	hs := http.Header{}

	for k, v := range headers {
		hs.Set(k, v)
	}

	return &Cloud{
		headers: hs,
	}
}

func (c *Cloud) Check(ctx context.Context) (check.Message, error) {
	if c.client == nil {
		c.client = cc.Client(ctx)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", cloudApi, nil)
	if err != nil {
		return nil, err
	}

	req.Header = c.headers

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code is %d", resp.StatusCode)
	}

	type respInfo struct {
		Retcode int    `json:"retcode"`
		Message string `json:"message"`
		Data    struct {
			FreeTime CloudFreeTime `json:"free_time"`
		} `json:"data"`
	}

	var info respInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if info.Message != "OK" {
		return nil, fmt.Errorf("message is %s", info.Message)
	}

	info.Data.FreeTime.Message = info.Message

	return info.Data.FreeTime, nil
}

type CloudFreeTime struct {
	Message string `json:"-"`

	SendFreetime  string `json:"send_freetime"`
	FreeTime      string `json:"free_time"`
	FreeTimeLimit string `json:"free_time_limit"`
	OverFreetime  string `json:"over_freetime"`
}

func (CloudFreeTime) Title() string {
	return "cloud game"
}

func (cft CloudFreeTime) Content() string {
	var fts, ftls string

	ft, err := time.ParseDuration(cft.FreeTime + "m")
	if err != nil {
		fts = cft.FreeTime
	} else {
		fts = ft.String()
	}

	ftl, err := time.ParseDuration(cft.FreeTimeLimit + "m")
	if err != nil {
		ftls = cft.FreeTimeLimit
	} else {
		ftls = ftl.String()
	}

	return fmt.Sprintf(
		`Ret: %s
Free: %s / %s`,
		cft.Message,
		fts,
		ftls,
	)
}
