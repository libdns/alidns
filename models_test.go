package alidns

import (
	"errors"
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

func TestOpError(t *testing.T) {
	type testCase struct {
		memo   string
		record error
		result error
	}

	cases := []testCase{
		{
			memo:   "error with op",
			record: OpError("op", errors.New("something wrong")),
			result: errors.New("there is something error when 'op': caused with: something wrong"),
		},
		{
			memo:   "error with none error",
			record: OpError("op",nil),
			result: nil,
		},
	}

	for _, c := range cases {
		if (c.record != nil && c.result != nil )&& c.record.Error() != c.result.Error() {
			t.Log("excepted:", c.result.Error(), "got:", c.record.Error())
			t.Fail()
			return
		}
		t.Log("case ", c.memo, "was pass.")
	}
}

func TestOpErrors(t *testing.T) {
	type testCase struct {
		memo   string
		record error
		result error
	}

	cases := []testCase{
		{
			memo:   "error with op",
			record: OpErrors("op").JoinError(errors.New("something wrong")).Error(),
			result: errors.New("there is something error when 'op': caused with: something wrong"),
		},
		{
			memo:   "error with op and record",
			record: OpErrors("op").JoinRecord(libdns.RR{Name: "rec"}, errors.New("something wrong")).Error(),
			result: errors.New("there is something error when 'op': caused at record named 'rec': something wrong"),
		},
		{
			memo:   "error with none error",
			record: OpErrors("op").Error(),
			result: nil,
		},
	}

	for _, c := range cases {
		if (c.record != nil && c.result != nil )&& c.record.Error() != c.result.Error() {
			t.Log("excepted:", c.result.Error(), "got:", c.record.Error())
			t.Fail()
			return
		}
		t.Log("case ", c.memo, "was pass.")
	}
}
