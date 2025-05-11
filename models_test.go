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
			result: aliDomaRecord{
				Rr:    "sub",
				DTyp:     "HTTPS",
				Priority: 100,
				DVal:     "target.com alpn=333",
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
