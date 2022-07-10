package ys

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/togls/gi-helper/check"
	cc "github.com/togls/gi-helper/context"
)

const (
	baseApi      = "https://api-takumi.mihoyo.com"
	rolesInfoApi = baseApi + "/binding/api/getUserGameRolesByCookie?game_biz=%s"
	signInfoApi  = baseApi + "/event/bbs_sign_reward/info?act_id=" + actID + "&uid=%s&region=%s"
	signApi      = baseApi + "/event/bbs_sign_reward/sign"
)

const (
	actID = "e202009291139501"

	appVersion = "2.3.0"
	userAgent  = "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) miHoYoBBS/" + appVersion
)

type YS struct {
	client *http.Client

	checks *checks
	cookie string
}

func New(cookie string) *YS {
	if cookie == "" {
		panic("cookie is empty")
	}

	return &YS{
		cookie: cookie,
	}
}

func (ys *YS) Check(ctx context.Context) (check.Message, error) {
	if ys.checks == nil {

		roles, err := ys.getRolesInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("get roles info: %w", err)
		}

		if len(roles) == 0 {
			return nil, fmt.Errorf("no roles")
		}

		ys.checks = new(checks)

		for _, role := range roles {
			si, err := ys.getSignInfo(ctx, role.GameUid, role.Region)
			if err != nil {
				return nil, fmt.Errorf("get sign info: %w", err)
			}

			role := role

			ys.checks.infos = append(ys.checks.infos, CheckInfo{
				RoleInfo: &role,
				SignInfo: si,
			})
		}
	}

	ys.checks.fails = 0

	isSign := func(t time.Time) bool {
		ny, nm, nd := time.Now().Date()
		y, m, d := t.Date()

		return ny == y && nm == m && nd == d
	}

	for _, role := range ys.checks.infos {
		role.SignInfo.Status = "✅ Signed"

		if !isSign(role.Today) &&
			!role.SignInfo.IsSign {

			role.SignInfo.Today = time.Now()

			status, err := ys.postSign(ctx, role.GameUid, role.Region)
			if err != nil {
				role.SignInfo.Status = "❌ Failed to check-in"
				ys.checks.fails++
				continue
			}

			role.SignInfo.IsSign = true
			role.SignInfo.TotalSignDay++
			role.SignInfo.Status = status
		}
	}

	return ys.checks, nil
}

type CheckInfo struct {
	*RoleInfo
	*SignInfo
}

type checks struct {
	infos []CheckInfo
	fails int32
}

func (cs *checks) Title() string {
	return "mihoyo bbs"
}

func (cs *checks) Content() string {
	content := new(bytes.Buffer)

	fmt.Fprintf(content,
		`☁️ ✔️ %d · ✖️ %d

	
	`,
		len(cs.infos)-int(cs.fails),
		cs.fails)

	for i, info := range cs.infos {
		fmt.Fprintf(content,
			`🌈No.%d
			####%s####
			🔅%s %d %s
			Total monthly check-ins: %d days
			Status: %s
			##################`,
			i+1,
			info.Today.Format("2006-01-02"),
			info.Nickname, info.Level, info.RegionName, info.TotalSignDay,
			info.Status,
		)
	}

	return content.String()
}

type RoleInfo struct {
	GameBiz    string `json:"game_biz"`
	Region     string `json:"region"`
	GameUid    string `json:"game_uid"`
	Nickname   string `json:"nickname"`
	Level      int    `json:"level"`
	IsChosen   bool   `json:"is_chosen"`
	RegionName string `json:"region_name"`
	IsOfficial bool   `json:"is_official"`
}

type SignInfo struct {
	TotalSignDay  int       `json:"total_sign_day"`
	Today         time.Time `json:"today"`
	IsSign        bool      `json:"is_sign"`
	FirstBind     bool      `json:"first_bind"`
	IsSub         bool      `json:"is_sub"`
	MonthFirst    bool      `json:"month_first"`
	SignCntMissed int       `json:"sign_cnt_missed"`

	Status string `json:"-"`
}

func (si *SignInfo) UnmarshalJSON(data []byte) error {
	type Alias SignInfo

	alias := struct {
		Today string `json:"today"`
		*Alias
	}{
		Alias: (*Alias)(si),
	}
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", alias.Today)
	if err != nil {
		return err
	}

	si.Today = t
	return nil
}

func (ys *YS) getRolesInfo(ctx context.Context) ([]RoleInfo, error) {
	if ys.client == nil {
		ys.client = cc.Client(ctx)
	}

	biz := "hk4e_cn"
	url := fmt.Sprintf(rolesInfoApi, biz)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = ys.headers(false)

	resp, err := ys.client.Do(req)
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
			List []RoleInfo `json:"list"`
		} `json:"data"`
	}

	var info respInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if info.Message != "OK" {
		return nil, fmt.Errorf("message is %s", info.Message)
	}

	return info.Data.List, nil
}

func (ys *YS) getSignInfo(ctx context.Context, uuid, region string) (*SignInfo, error) {
	if ys.client == nil {
		ys.client = cc.Client(ctx)
	}

	url := fmt.Sprintf(signInfoApi, uuid, region)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = ys.headers(false)

	resp, err := ys.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code is %d", resp.StatusCode)
	}

	type respInfo struct {
		Retcode int      `json:"retcode"`
		Message string   `json:"message"`
		Data    SignInfo `json:"data"`
	}

	var info respInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if info.Message != "OK" {
		return nil, fmt.Errorf("message is %s", info.Message)
	}

	return &info.Data, nil
}

func (ys *YS) postSign(ctx context.Context, uid, region string) (string, error) {
	if ys.client == nil {
		ys.client = cc.Client(ctx)
	}

	payload := map[string]string{
		"act_id": actID,
		"uid":    uid,
		"region": region,
	}

	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", signApi, buf)
	if err != nil {
		return "", err
	}

	req.Header = ys.headers(true)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ys.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 0:      success
	// -5003:  already checked in

	type respInfo struct {
		Retcode int    `json:"retcode"`
		Message string `json:"message"`
	}

	var info respInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	if info.Retcode == 0 ||
		info.Retcode == -5003 {
		return info.Message, nil
	}

	return "", fmt.Errorf("sign err, retcode=%d, message=%s", info.Retcode, info.Message)
}

func (ys *YS) headers(withDS bool) http.Header {
	header := http.Header{}
	header.Set("User-Agent", userAgent)
	header.Set("Cookie", ys.cookie)

	if withDS {
		u := uuid.NewMD5(uuid.NameSpaceURL, []byte(userAgent))
		deviceID := strings.ReplaceAll(u.String(), "-", "")

		header.Set("x-rpc-device_id", deviceID)
		header.Set("x-rpc-client_type", "5")
		header.Set("x-rpc-app_version", appVersion)
		header.Set("DS", ds())
	}

	return header
}

func ds() string {
	salt := "h8w582wxwgqvahcdkpvdhbh2w9casgfl"

	t := time.Now().Unix()
	r := randString(6)

	hash := md5.Sum([]byte(fmt.Sprintf("salt=%s&t=%d&r=%s", salt, t, r)))

	return fmt.Sprintf("%d,%s,%s", t, r, hex.EncodeToString(hash[:]))
}

func randString(n int) string {
	sample := []byte("abcdefghijklmnopqrstuvwxyz1234567890")

	b := make([]byte, n)
	for i := range b {
		b[i] = sample[rand.Intn(len(sample))]
	}

	return string(b)
}
