package alidns

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultRegionID string = "cn-hangzhou"
const addressOfAPI string = "%s://alidns.aliyuncs.com/"

// CredentialInfo implements param of the crediential
type CredentialInfo struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	RegionID        string `json:"region_id,omitempty"`
	SecurityToken   string `json:"security_token,omitempty"`
}

// aliClientSchema abstructs the alidns.Client
type aliClientSchema struct {
	mutex        sync.Mutex
	APIHost      string
	requestMap   []keyPair
	signString   string
	signPassword string
}

// keyPair implments of K-V struct
type keyPair struct {
	Key   string
	Value string
}

func NewCredentialInfo(accessKeyID, accessKeySecret, regionID string) *CredentialInfo {
	if accessKeyID == "" || accessKeySecret == "" {
		return nil
	}
	if len(regionID) == 0 {
		regionID = defaultRegionID
	}
	return &CredentialInfo{
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		RegionID:        regionID,
	}
}

func getClientSchema(cred *CredentialInfo, scheme string) (*aliClientSchema, error) {
	if cred == nil {
		return &aliClientSchema{}, errors.New("alidns: credentials missing")
	}
	if scheme == "" {
		scheme = "http"
	}

	cl0 := &aliClientSchema{
		APIHost: fmt.Sprintf(addressOfAPI, scheme),
		requestMap: []keyPair{
			{Key: "AccessKeyId", Value: cred.AccessKeyID},
			{Key: "Format", Value: "JSON"},
			{Key: "SignatureMethod", Value: "HMAC-SHA1"},
			{Key: "SignatureNonce", Value: fmt.Sprintf("%d", time.Now().UnixNano())},
			{Key: "SignatureVersion", Value: "1.0"},
			{Key: "Timestamp", Value: time.Now().UTC().Format("2006-01-02T15:04:05Z")},
			{Key: "Version", Value: "2015-01-09"},
		},
		signString:   "",
		signPassword: cred.AccessKeySecret,
	}

	return cl0, nil
}

func (c *aliClientSchema) signReq(method string) error {
	if c.signPassword == "" || len(c.requestMap) == 0 {
		return errors.New("alidns: AccessKeySecret or Request(includes AccessKeyId) is Misssing")
	}
	sort.Sort(byKey(c.requestMap))
	str := c.reqMapToStr()
	str = c.reqStrToSign(str, method)
	c.signString = signStr(str, c.signPassword)
	return nil
}

func (c *aliClientSchema) addReqBody(key string, value string) error {
	if key == "" && value == "" {
		return errors.New("key or value is Empty")
	}
	el := keyPair{Key: key, Value: value}
	c.mutex.Lock()
	for _, el0 := range c.requestMap {
		if el.Key == el0.Key {
			c.mutex.Unlock()
			return errors.New("duplicate keys")
		}
	}
	c.requestMap = append(c.requestMap, el)
	c.mutex.Unlock()
	return nil
}

func (c *aliClientSchema) setReqBody(key string, value string) error {
	if key == "" && value == "" {
		return errors.New("key or value is Empty")
	}
	el := keyPair{Key: key, Value: value}
	c.mutex.Lock()
	for in, el0 := range c.requestMap {
		if el.Key == el0.Key {
			(c.requestMap)[in] = el
			c.mutex.Unlock()
			return nil
		}
	}
	c.mutex.Unlock()
	return fmt.Errorf("entry of %s not found", key)
}

func (c *aliClientSchema) reqStrToSign(src string, method string) string {
	if method == "" {
		method = "GET"
	}
	ecReq := urlEncode(src)
	return fmt.Sprintf("%s&%s&%s", method, "%2F", ecReq)
}

func (c *aliClientSchema) reqMapToStr() string {
	m0 := c.requestMap
	urlEn := url.Values{}
	c.mutex.Lock()
	for _, o := range m0 {
		urlEn.Add(o.Key, o.Value)
	}
	c.mutex.Unlock()
	return urlEn.Encode()
}

// HttpRequest generates http.Request from schema
func (c *aliClientSchema) HttpRequest(cxt context.Context, method string, body io.Reader) (*http.Request, error) {
	if method == "" {
		method = "GET"
	}
	c.signReq(method)
	si0 := fmt.Sprintf("%s=%s", "Signature", strings.ReplaceAll(c.signString, "+", "%2B"))
	mURL := fmt.Sprintf("%s?%s&%s", c.APIHost, c.reqMapToStr(), si0)
	req, err := http.NewRequestWithContext(cxt, method, mURL, body)
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return &http.Request{}, err
	}
	return req, nil
}

func signStr(src string, secret string) string {
	secret = secret + "&"
	hm := hmac.New(sha1.New, []byte(secret))
	hm.Write([]byte(src))
	sum := hm.Sum(nil)
	return base64.StdEncoding.EncodeToString(sum)
}

func goVer() float64 {
	versionString := runtime.Version()
	versionString, _ = strings.CutPrefix(versionString, "go")
	verStrings := strings.Split(versionString, ".")
	var result float64
	for i, v := range verStrings {
		tmp, _ := strconv.ParseFloat(v, 32)
		result = tmp * (1 / math.Pow10(i))
	}
	return result
}

func urlEncode(src string) string {
	str0 := src
	str0 = strings.Replace(str0, "+", "%20", -1)
	str0 = strings.Replace(str0, "*", "%2A", -1)
	str0 = strings.Replace(str0, "%7E", "~", -1)

	str0 = url.QueryEscape(str0)
	if goVer() > 1.20 {
		str0 = strings.Replace(str0, "%26", "&", -1)
	}

	return str0
}

type byKey []keyPair

func (v byKey) Len() int {
	return len(v)
}

func (v byKey) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v byKey) Less(i, j int) bool {
	return v[i].Key < v[j].Key
}
