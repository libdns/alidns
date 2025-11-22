package alidns

import (
	"context"
	"fmt"
	"testing"
)

const AccessKeyID = "<Input your AccessKeyID here>"
const AccessKeySecret = "<Input your AccessKeySecret here>"

func Test_URLEncode(t *testing.T) {
	s0 := urlEncode("AccessKeyId=testid&Action=DescribeDomainRecords")
	if s0 != "AccessKeyId%3Dtestid%26Action%3DDescribeDomainRecords" {
		t.Log(s0)
		t.Fail()
	}
	t.Log(s0)
}

var cl0 = &aliClientSchema{
	APIHost: fmt.Sprintf(addressOfAPI, "https"),
	requestMap: []keyPair{
		{Key: "AccessKeyId", Value: "testid"},
		{Key: "Format", Value: "XML"},
		{Key: "Action", Value: "DescribeDomainRecords"},
		{Key: "SignatureMethod", Value: "HMAC-SHA1"},
		{Key: "DomainName", Value: "example.com"},
		{Key: "SignatureVersion", Value: "1.0"},
		{Key: "SignatureNonce", Value: "f59ed6a9-83fc-473b-9cc6-99c95df3856e"},
		{Key: "Timestamp", Value: "2016-03-24T16:41:54Z"},
		{Key: "Version", Value: "2015-01-09"},
	},
	signString:   "",
	signPassword: "testsecret",
}

func Test_AliClintReq(t *testing.T) {
	str := cl0.reqMapToStr()
	t.Log("map to str:" + str + "\n")
	str = cl0.reqStrToSign(str, "GET")

	// validate sign string from doc: https://help.aliyun.com/document_detail/29747.html#:~:text=%E9%82%A3%E4%B9%88-,stringtosign
	if goVer() > 1.20 && str != "GET&%2F&AccessKeyId%3Dtestid&Action%3DDescribeDomainRecords&DomainName%3Dexample.com&Format%3DXML&SignatureMethod%3DHMAC-SHA1&SignatureNonce%3Df59ed6a9-83fc-473b-9cc6-99c95df3856e&SignatureVersion%3D1.0&Timestamp%3D2016-03-24T16%253A41%253A54Z&Version%3D2015-01-09" {
		t.Error("sign str error")
	}
	if goVer() < 1.20 && str != "GET&%2F&AccessKeyId%3Dtestid%26Action%3DDescribeDomainRecords%26DomainName%3Dexample.com%26Format%3DXML%26SignatureMethod%3DHMAC-SHA1%26SignatureNonce%3Df59ed6a9-83fc-473b-9cc6-99c95df3856e%26SignatureVersion%3D1.0%26Timestamp%3D2016-03-24T16%253A41%253A54Z%26Version%3D2015-01-09" {
		t.Error("sign str error")
	}
	t.Log("sign str:" + str + "\n")
	t.Log("signed base64:" + signStr(str, cl0.signPassword) + "\n")

}

func Test_AppendDupReq(t *testing.T) {
	err := cl0.addReqBody("Version", "100")
	if err == nil {
		t.Fail()
	}
}

var p0 = Provider{
	AccKeyID:     AccessKeyID,
	AccKeySecret: AccessKeySecret,
}

func Test_RequestUrl(t *testing.T) {
	p0.getClient()
	p0.client.AddRequestBody("Action", "DescribeDomainRecords")
	p0.client.AddRequestBody("DomainName", "viscrop.top")
	p0.client.SetRequestBody("Timestamp", "2020-10-16T20:10:54Z")
	r, err := p0.client.schema.HttpRequest(context.TODO(), "GET", nil)
	t.Log("url:", r.URL.String(), "err:", err)
}
