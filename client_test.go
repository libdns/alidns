package alidns

import (
	"context"
	"testing"

	"github.com/libdns/libdns"
)

func Test_ClientAPIReq(t *testing.T) {
	p0.getClient()
	p0.client.aClient.addReqBody("Action", "DescribeDomainRecords")
	p0.client.aClient.addReqBody("KeyWords", "vi")
	var rs aliDomaRecords
	rspData := aliResult{}
	err := p0.doAPIRequest(context.TODO(), &rspData)
	t.Log("req", p0.client.aClient, "data", rspData, "err:", err, "rs:", rs)
}

func Test_QueryDomainRecord(t *testing.T) {
	rr, name, _ := p0.queryMainDomain(context.Background(), "www.viscrop.top")
	r0, err := p0.queryDomainRecord(context.TODO(), rr, name, "A")
	t.Log("result:", r0, "err:", err)
	r0, err = p0.queryDomainRecord(context.TODO(), rr, name, "A", ".")
	t.Log("result with A rec:", r0, "err:", err)
}

func Test_QueryDomainRecords(t *testing.T) {
	_, name, _ := p0.queryMainDomain(context.Background(), "me.viscrop.top")
	r0, err := p0.queryDomainRecords(context.TODO(), name)
	t.Log("result:", r0, "err:", err)
	_, name, _ = p0.queryMainDomain(context.Background(), "me.viscraop.top")
	r0, err = p0.queryDomainRecords(context.TODO(), name)
	t.Log("result:", r0, "err:", err)
}

func Test_DomainRecordOp(t *testing.T) {
	dr0 := aliDomaRecord{
		DName: "viscrop.top",
		Rr:    "baidu",
		DTyp:  "CNAME",
		DVal:  "baidu.com",
		TTL:   600,
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

func Test_aliDomainRecordWithZone(t *testing.T) {
	type testCase struct {
		memo   string
		record libdns.Record
		zone   string
		result aliDomaRecord
	}

	cases := []testCase{
		{
			memo: "record.Name without zone",
			record: libdns.RR{
				Name: "sub",
			},
			zone: "mydomain.com.",
			result: aliDomaRecord{
				Rr:    "sub",
				DName: "mydomain.com",
			},
		},
		{
			memo: "record.Name with zone",
			record: libdns.RR{
				Name: "sub.mydomain.com",
			},
			zone: "mydomain.com.",
			result: aliDomaRecord{
				Rr:    "sub",
				DName: "mydomain.com",
			},
		},
	}

	for _, c := range cases {
		rec := alidnsRecord(c.record, c.zone)
		if !rec.Equals(c.result) {
			t.Log("excepted:", c.result, "got:", rec)
			t.Fail()
		}
		t.Log("case ", c.memo, "was pass.")
	}

}

func Test_aliDomainRecord(t *testing.T) {
	type testCase struct {
		memo   string
		record libdns.Record
		result aliDomaRecord
	}

	cases := []testCase{
		{
			memo: "normal record",
			record: libdns.RR{
				Name: "sub",
				Type: "A",
				Data: "1.1.1.1",
			},
			result: aliDomaRecord{
				Rr:   "sub",
				DTyp: "A",
				DVal: "1.1.1.1",
			},
		},
	}

	for _, c := range cases {
		rec := alidnsRecord(c.record)
		if !rec.Equals(c.result) {
			t.Log("excepted:", c.result, "got:", rec)
			t.Fail()
		}
		t.Log("case ", c.memo, "was pass.")
	}

}
