package alidns

import (
	"context"
	"testing"
)

func Test_ClientAPIReq(t *testing.T) {
	p0.getClient()
	p0.client.SetRequestBody("Action", "DescribeDomainRecords")
	p0.client.SetRequestBody("KeyWords", "vi")
	var rs aliDomaRecords
	rspData := aliDomainResult{}
	err := p0.client.doAPIRequest(context.TODO(), &rspData)
	t.Log("req", p0.client.schema, "data", rspData, "err:", err, "rs:", rs)
}

func Test_QueryDomainRecord(t *testing.T) {
	r0, err := p0.queryDomainRecord(context.TODO(), "*", "aliyun.viscropst.ren", "A")
	t.Log("result:", r0)
	if err != nil {
		t.Log("err:", err)
		t.Fail()
	}
	t.Log("result with A rec:", r0, "err:", err)
}

func Test_QueryDomainRecords(t *testing.T) {
	r0, err := p0.queryDomainRecords(context.TODO(), "me.viscrop.top")
	t.Log("result:", r0, "err:", err)
}

func Test_DomainRecordOp(t *testing.T) {
	dr0 := aliDomainRecord{
		DomainName:  "viscrop.top",
		Rr:          "baidu",
		DomainType:  "CNAME",
		DomainValue: "baidu.com",
		TTL:         600,
	}
	r0, err := p0.addDomainRecord(context.TODO(), dr0)
	t.Log("result:", r0, "err:", err)
	dr0, err = p0.getDomainRecord(context.TODO(), r0)
	t.Log("result:", dr0, "err:", err)
	dr0.Rr = "bai"
	r0, err = p0.setDomainRecord(context.TODO(), dr0)
	t.Log("result:", r0, "err:", err)
	r0, err = p0.delDomainRecord(context.TODO(), dr0)
	t.Log("result:", r0, "err:", err)
}
