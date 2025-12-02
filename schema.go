package alidns

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
	headerPairs  KeyPairs
	requestPairs KeyPairs
	signString   string
	signPassword string
	version      int
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
	return defaultSchemaV2(cred, scheme)
}

func (c *aliClientSchema) signReq(method string) error {
	if c.signPassword == "" || len(c.requestPairs) == 0 {
		return errors.New("alidns: AccessKeySecret or Request(includes AccessKeyId) is Misssing")
	}
	if c.version == 2 {
		return c.signReqV2(method)
	}
	return nil
}

func (c *aliClientSchema) addReqBody(key string, value string) error {
	c.mutex.Lock()
	var err error
	c.requestPairs, err = c.requestPairs.Append(key, value)
	if err != nil {
		c.mutex.Unlock()
		return err
	}
	c.mutex.Unlock()
	return nil
}

func (c *aliClientSchema) setReqBody(key string, value string) error {
	c.mutex.Lock()
	err := c.requestPairs.Update(key, value)
	if err != nil {
		c.mutex.Unlock()
		return err
	}
	c.mutex.Unlock()
	return nil
}

func (c *aliClientSchema) reqMapToStr() string {
	c.mutex.Lock()
	result := c.requestPairs.UrlEncodedString()
	c.mutex.Unlock()
	return result
}

// HttpRequest generates http.Request from schema
func (c *aliClientSchema) HttpRequest(cxt context.Context, method string) (*http.Request, error) {
	if method == "" {
		method = http.MethodGet
	}
	c.signReq(method)
	requestUrl := c.APIHost
	if method == http.MethodGet {
		requestUrl = fmt.Sprintf("%s?%s", requestUrl, c.reqMapToStr())
	}
	if c.version == 2 {
		requestUrl = c.urlV2(requestUrl)
	}
	var bodyReader io.Reader
	if method == http.MethodPost && c.version > 2 {
		c.headerPairs.Append("content-type", "application/x-www-form-urlencoded")
		bodyReader = bytes.NewReader([]byte(c.reqMapToStr()))
	}
	req, err := http.NewRequestWithContext(cxt, method, requestUrl, bodyReader)
	if err != nil {
		return &http.Request{}, err
	}
	req.Header.Set("Accept", "application/json")
	if len(c.headerPairs) == 0 {
		return req, nil
	}
	for _, v := range c.headerPairs {
		req.Header.Set(v.Key, v.Value)
	}
	return req, nil
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

func percentCode(src string) string {
	result := strings.Replace(src, "+", "%20", -1)
	result = strings.Replace(result, "*", "%2A", -1)
	result = strings.Replace(result, "%7E", "~", -1)
	return result
}

func urlEncode(src string) string {
	str0 := percentCode(src)
	str0 = url.QueryEscape(str0)
	if goVer() > 1.20 {
		str0 = strings.Replace(str0, "%26", "&", -1)
	}

	return str0
}

type KeyPairs []keyPair

func (p KeyPairs) Append(key, value string) (KeyPairs, error) {
	if key == "" && value == "" {
		return p, errors.New("key or value is Empty")
	}
	srcEl := keyPair{Key: key, Value: value}
	for _, el := range p {
		if srcEl.Key == el.Key {
			return p, errors.New("duplicate keys")
		}
	}
	p = append(p, srcEl)
	return p, nil
}

func (p KeyPairs) Update(key, value string) error {
	if key == "" && value == "" {
		return errors.New("key or value is Empty")
	}
	srcEl := keyPair{Key: key, Value: value}
	for in, el := range p {
		if srcEl.Key == el.Key {
			p[in] = srcEl
			return nil
		}
	}
	return fmt.Errorf("entry of %s not found", key)
}

func (p KeyPairs) UrlEncodedString() string {
	urlEn := url.Values{}
	for _, o := range p {
		urlEn.Add(o.Key, o.Value)
	}
	return urlEn.Encode()
}

func (p KeyPairs) Len() int {
	return len(p)
}

func (p KeyPairs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p KeyPairs) Less(i, j int) bool {
	return p[i].Key < p[j].Key
}
