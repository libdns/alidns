package alidns

import (
	"testing"

	"github.com/libdns/libdns"
)

func Test_aliDomainRecordWithZone(t *testing.T) {
	type testCase struct {
		memo   string
		record libdns.Record
		zone   string
		result aliDomainRecord
	}

	cases := []testCase{
		{
			memo: "record.Name without zone",
			record: libdns.RR{
				Name: "sub",
			},
			zone: "mydomain.com.",
			result: aliDomainRecord{
				Rr:         "sub",
				DomainName: "mydomain.com",
			},
		},
		{
			memo: "record.Name with zone",
			record: libdns.RR{
				Name: "sub.mydomain.com",
			},
			zone: "mydomain.com.",
			result: aliDomainRecord{
				Rr:         "sub",
				DomainName: "mydomain.com",
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
		result aliDomainRecord
	}

	cases := []testCase{
		{
			memo: "normal record",
			record: libdns.RR{
				Name: "sub",
				Type: "A",
				Data: "1.1.1.1",
			},
			result: aliDomainRecord{
				Rr:          "sub",
				DomainType:  "A",
				DomainValue: "1.1.1.1",
			},
		},
		{
			memo: "HTTPS record",
			record: libdns.ServiceBinding{
				Name:     "sub",
				Scheme:   "https",
				Target:   "target.com",
				Priority: 100,
				Params: map[string][]string{
					"alpn": {"333"},
				},
			},
			result: aliDomainRecord{
				Rr:          "sub",
				DomainType:  "HTTPS",
				Priority:    100,
				DomainValue: "target.com alpn=333",
			},
		},
	}

	for _, c := range cases {
		rec := alidnsRecord(c.record)
		if !rec.Equals(c.result) {
			t.Log("excepted:", c.result, "got:", rec)
			t.Fail()
			return
		}
		t.Log("case ", c.memo, "was pass.")
	}

}
