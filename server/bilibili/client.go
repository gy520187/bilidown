package bilibili

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"sync"
	"time"

	"bilidown/util"
)

type BiliClient struct {
	SESSDATA string
	Jar      http.CookieJar
	mixinKey string
	mixinAt  time.Time
}

type qrLoginSession struct {
	jar http.CookieJar
	at  time.Time
}

var qrLoginSessions sync.Map

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// SimpleGET 简单的 GET 请求
func (client *BiliClient) SimpleGET(_url string, params map[string]string) (*http.Response, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return client.doGET(_url, values)
}

// SignedGET 带 WBI 签名的 GET 请求
func (client *BiliClient) SignedGET(_url string, params map[string]string) (*http.Response, error) {
	mixinKey, err := client.ensureMixinKey()
	if err != nil {
		return nil, err
	}
	return client.doGET(_url, WbiSign(params, mixinKey))
}

func (client *BiliClient) doGET(_url string, values url.Values) (*http.Response, error) {
	query := ""
	if len(values) > 0 {
		query = "?" + values.Encode()
	}
	_client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(nil),
		},
		Jar: client.Jar,
	}
	request, err := http.NewRequest("GET", _url+query, nil)
	if err != nil {
		return nil, err
	}
	request.Header = client.MakeHeader()
	return _client.Do(request)
}

// MakeHeader 生成请求头
func (client *BiliClient) MakeHeader() http.Header {
	header := http.Header{}
	if client.SESSDATA != "" {
		header.Set("Cookie", "SESSDATA="+client.SESSDATA)
	}
	header.Set("User-Agent", defaultUserAgent)
	header.Set("Referer", "https://www.bilibili.com/")
	header.Set("Origin", "https://www.bilibili.com")
	return header
}

// CheckLogin 检查是否已经登录
func (client *BiliClient) CheckLogin() (bool, error) {
	response, err := client.SimpleGET("https://api.bilibili.com/x/space/myinfo", nil)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	body := BaseResV2{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return false, err
	}
	if body.Code != 0 {
		return false, errors.New(body.Message)
	}
	return body.Success(), nil
}

// NewQRInfo 获取登录二维码信息
func (client *BiliClient) NewQRInfo() (*QRInfo, error) {
	response, err := client.SimpleGET("https://passport.bilibili.com/x/passport-login/web/qrcode/generate", map[string]string{
		"source": "main-fe-header",
	})
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body := BaseResV2{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return nil, err
	}
	if body.Code != 0 {
		return nil, errors.New(body.Message)
	}
	qrInfo := QRInfo{}
	err = json.Unmarshal(body.Data, &qrInfo)
	if err != nil {
		return nil, err
	}
	if qrInfo.QrcodeKey != "" && client.Jar != nil {
		qrLoginSessions.Store(qrInfo.QrcodeKey, &qrLoginSession{jar: client.Jar, at: time.Now()})
	}
	return &qrInfo, nil
}

func NewQRJar() http.CookieJar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}
	return jar
}

func loadQRJar(qrKey string) http.CookieJar {
	purgeExpiredQRSessions()
	value, ok := qrLoginSessions.Load(qrKey)
	if !ok {
		return NewQRJar()
	}
	session, ok := value.(*qrLoginSession)
	if !ok || session == nil || session.jar == nil {
		return NewQRJar()
	}
	return session.jar
}

func purgeExpiredQRSessions() {
	cutoff := time.Now().Add(-10 * time.Minute)
	qrLoginSessions.Range(func(key, value any) bool {
		session, ok := value.(*qrLoginSession)
		if !ok || session == nil || session.at.Before(cutoff) {
			qrLoginSessions.Delete(key)
		}
		return true
	})
}

func (client *BiliClient) getWbiKeyRemote() (wbiKey string, err error) {
	response, err := client.SimpleGET("https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body := BaseResV2{}
	if err = json.NewDecoder(response.Body).Decode(&body); err != nil {
		return "", err
	}
	var data struct {
		WbiImg struct {
			ImgURL string `json:"img_url"`
			SubURL string `json:"sub_url"`
		} `json:"wbi_img"`
	}
	if err = json.Unmarshal(body.Data, &data); err != nil {
		return "", err
	}
	match := regexp.MustCompile(`/bfs/wbi/([a-z0-9]+)\.`)
	imgMatch := match.FindStringSubmatch(data.WbiImg.ImgURL)
	subMatch := match.FindStringSubmatch(data.WbiImg.SubURL)
	if len(imgMatch) < 2 || len(subMatch) < 2 {
		return "", errors.New("解析 wbi key 失败")
	}
	return imgMatch[1] + subMatch[1], nil
}

// GetQRStatus 获取二维码状态
func (client *BiliClient) GetQRStatus(qrKey string) (qrStatus *QRStatus, sessdata string, err error) {
	if client.Jar == nil {
		client.Jar = loadQRJar(qrKey)
	}
	params := map[string]string{
		"qrcode_key": qrKey,
		"source":     "main-fe-header",
	}
	response, err := client.SimpleGET("https://passport.bilibili.com/x/passport-login/web/qrcode/poll", params)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	body := BaseResV2{}
	err = json.NewDecoder(response.Body).Decode(&body)
	if err != nil {
		return nil, "", err
	}
	if body.Code != 0 {
		return nil, "", errors.New(body.Message)
	}
	qrStatus = &QRStatus{}
	err = json.Unmarshal(body.Data, &qrStatus)
	if err != nil {
		return nil, "", err
	}
	if qrStatus.Code != 0 {
		return qrStatus, "", nil
	}
	sessdata, err = GetCookieValue(response.Cookies(), "SESSDATA")
	if err != nil && client.Jar != nil {
		if parsed, parseErr := url.Parse("https://passport.bilibili.com"); parseErr == nil {
			sessdata, err = GetCookieValue(client.Jar.Cookies(parsed), "SESSDATA")
		}
	}
	if err != nil {
		return nil, "", err
	}
	qrLoginSessions.Delete(qrKey)
	return qrStatus, sessdata, nil
}

// GetCookieValue 获取指定 Name 的 Cookie 值
func GetCookieValue(cookies []*http.Cookie, name string) (string, error) {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value, nil
		}
	}
	return "", errors.New("cookie with name " + name + " not found")
}

// SaveSessdata 保存 SESSDATA
func SaveSessdata(db *sql.DB, sessdata string) error {
	util.SqliteLock.Lock()
	_, err := db.Exec(`INSERT OR REPLACE INTO "field" ("name", "value") VALUES ("sessdata", ?)`, sessdata)
	util.SqliteLock.Unlock()
	return err
}

// GetSessdata 获取 SESSDATA
func GetSessdata(db *sql.DB) (string, error) {
	util.SqliteLock.Lock()
	row := db.QueryRow(`SELECT "value" FROM "field" WHERE "name" = "sessdata"`)
	util.SqliteLock.Unlock()
	var sessdata string
	err := row.Scan(&sessdata)
	return sessdata, err
}
