package alidns

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type ttl_t = uint32

type aliDomainRecord struct {
	Rr          string `json:"RR,omitempty"`
	Line        string `json:"Line,omitempty"`
	Status      string `json:"Status,omitempty"`
	Locked      bool   `json:"Locked,omitempty"`
	DomainType  string `json:"Type,omitempty"`
	DomainName  string `json:"DomainName,omitempty"`
	DomainValue string `json:"Value,omitempty"`
	RecordID    string `json:"RecordId,omitempty"`
	TTL         ttl_t  `json:"TTL,omitempty"`
	Weight      int    `json:"Weight,omitempty"`
	Priority    ttl_t  `json:"Priority,omitempty"`
}

func (r aliDomainRecord) RR() libdns.RR {
	return libdns.RR{
		Type: r.DomainType,
		Name: r.Rr,
		Data: r.DomainValue,
		TTL:  time.Duration(r.TTL) * time.Second,
	}
}

func (r aliDomainRecord) Equals(v aliDomainRecord) bool {
	result := v.Rr == r.Rr
	result = result && v.DomainName == r.DomainName
	result = result && v.DomainValue == r.DomainValue
	result = result && v.DomainType == r.DomainType
	return result
}

type aliDomaRecords struct {
	Record []aliDomainRecord `json:"Record,omitempty"`
}

type aliDomainResult struct {
	ReqID         string         `json:"RequestId,omitempty"`
	DomainRecords aliDomaRecords `json:"DomainRecords,omitempty"`
	DomainLevel   int            `json:"DomainLevel,omitempty"`
	DomainValue   string         `json:"Value,omitempty"`
	DomainName    string         `json:"DomainName,omitempty"`
	DomainType    string         `json:"Type,omitempty"`
	Rr            string         `json:"RR,omitempty"`
	TTL           ttl_t          `json:"TTL,omitempty"`
	Msg           string         `json:"Message,omitempty"`
	Rcmd          string         `json:"Recommend,omitempty"`
	HostID        string         `json:"HostId,omitempty"`
	Code          string         `json:"Code,omitempty"`
	TotalCount    int            `json:"TotalCount,omitempty"`
	PgSize        int            `json:"PageSize,omitempty"`
	PgNum         int            `json:"PageNumber,omitempty"`
	RecID         string         `json:"RecordId,omitempty"`
	Line          string         `json:"Line,omitempty"`
	Status        string         `json:"Status,omitempty"`
	Locked        bool           `json:"Locked,omitempty"`
	Weight        int            `json:"Weight,omitempty"`
	MinTTL        int            `json:"MinTtl,omitempty"`
	Priority      ttl_t          `json:"Priority,omitempty"`
}

func (r *aliDomainResult) ToDomaRecord() aliDomainRecord {
	return aliDomainRecord{
		RecordID:    r.RecID,
		DomainType:  r.DomainType,
		Rr:          r.Rr,
		DomainName:  r.DomainName,
		DomainValue: r.DomainValue,
		TTL:         r.TTL,
		Line:        r.Line,
		Status:      r.Status,
		Locked:      r.Locked,
		Weight:      r.Weight,
		Priority:    r.Priority,
	}
}

// AlidnsRecord convert libdns.Record with zone to aliDomaRecord
func alidnsRecord(r libdns.Record, zone ...string) aliDomainRecord {
	result := aliDomainRecord{}
	if r == nil {
		return result
	}
	tmpRR := r.RR()
	if len(zone) > 0 && len(zone[0]) > 0 {
		tmpZone := zone[0]
		result.Rr = libdns.RelativeName(tmpRR.Name, tmpZone)
		result.DomainName = strings.Trim(tmpZone, ".")
	} else {
		result.Rr = tmpRR.Name
	}
	result.DomainType = tmpRR.Type
	result.DomainValue = tmpRR.Data
	result.TTL = ttl_t(tmpRR.TTL.Seconds())
	if svcb, svcbok := r.(libdns.ServiceBinding); svcbok {
		result.Priority = ttl_t(svcb.Priority)
		result.DomainValue = fmt.Sprintf("%s %s", svcb.Target, svcb.Params)
	}
	return result
}

type opErrors struct {
	Op            string
	length        uint64
	errMsgBuilder strings.Builder
}

func OpErrors(op string) *opErrors {
	return &opErrors{
		Op:            op,
		errMsgBuilder: strings.Builder{},
		length:        0,
	}
}

func OpError(op string, err error) error {
	errs := OpErrors(op)
	errs.JoinError(err)
	return errs.Error()
}

func (e *opErrors) JoinError(err error) *opErrors {
	if err != nil {
		msg := "caused with: "
		msg += err.Error()
		e.errMsgBuilder.WriteString(msg + ",")
		e.length += 1
	}
	return e
}

func (e *opErrors) JoinRecord(record libdns.Record, err error) *opErrors {
	if err != nil {
		msg := "caused at record named "
		msg += "'" + record.RR().Name + "'"
		msg += ": "
		msg += err.Error()
		e.errMsgBuilder.WriteString(msg + ",")
		e.length += 1
	}
	return e
}

func (e *opErrors) errorMsg() string {
	msg := "there is something error "
	if len(e.Op) > 0 {
		msg += "when '" + e.Op + "'"
	}
	msg += ": "
	return strings.TrimSuffix(msg+e.errMsgBuilder.String(), ",")
}

func (e *opErrors) Error() error {
	if e.length > 0 {
		return errors.New(e.errorMsg())
	}
	return nil
}
