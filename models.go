package alidns

import (
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type ttl_t = uint32

type aliDomaRecord struct {
	Rr       string `json:"RR,omitempty"`
	Line     string `json:"Line,omitempty"`
	Status   string `json:"Status,omitempty"`
	Locked   bool   `json:"Locked,omitempty"`
	DTyp     string `json:"Type,omitempty"`
	DName    string `json:"DomainName,omitempty"`
	DVal     string `json:"Value,omitempty"`
	RecID    string `json:"RecordId,omitempty"`
	TTL      ttl_t  `json:"TTL,omitempty"`
	Weight   int    `json:"Weight,omitempty"`
	Priority ttl_t  `json:"Priority,omitempty"`
}

type aliDomaRecords struct {
	Record []aliDomaRecord `json:"Record,omitempty"`
}

type aliResult struct {
	ReqID      string         `json:"RequestId,omitempty"`
	DRecords   aliDomaRecords `json:"DomainRecords,omitempty"`
	DLvl       int            `json:"DomainLevel,omitempty"`
	DVal       string         `json:"Value,omitempty"`
	TTL        ttl_t          `json:"TTL,omitempty"`
	DName      string         `json:"DomainName,omitempty"`
	Rr         string         `json:"RR,omitempty"`
	Msg        string         `json:"Message,omitempty"`
	Rcmd       string         `json:"Recommend,omitempty"`
	HostID     string         `json:"HostId,omitempty"`
	Code       string         `json:"Code,omitempty"`
	TotalCount int            `json:"TotalCount,omitempty"`
	PgSize     int            `json:"PageSize,omitempty"`
	PgNum      int            `json:"PageNumber,omitempty"`
	DTyp       string         `json:"Type,omitempty"`
	RecID      string         `json:"RecordId,omitempty"`
	Line       string         `json:"Line,omitempty"`
	Status     string         `json:"Status,omitempty"`
	Locked     bool           `json:"Locked,omitempty"`
	Weight     int            `json:"Weight,omitempty"`
	MinTTL     int            `json:"MinTtl,omitempty"`
	Priority   ttl_t          `json:"Priority,omitempty"`
}

func (r *aliDomaRecord) LibdnsRecord() libdns.Record {
	return libdns.Record{
		ID:    r.RecID,
		Type:  r.DTyp,
		Name:  r.Rr,
		Value: r.DVal,
		TTL:   time.Duration(r.TTL) * time.Second,
	}
}

func (r *aliResult) ToDomaRecord() aliDomaRecord {
	return aliDomaRecord{
		RecID:  r.RecID,
		DTyp:   r.DTyp,
		Rr:     r.Rr,
		DName:  r.DName,
		DVal:   r.DVal,
		TTL:    r.TTL,
		Line:   r.Line,
		Status: r.Status,
		Locked: r.Locked,
		Weight: r.Weight,
	}
}

// AlidnsRecord convert libdns.Record to aliDomaRecord
func alidnsRecord(r libdns.Record) aliDomaRecord {
	result := aliDomaRecord{
		Rr:    r.Name,
		DTyp:  r.Type,
		DVal:  r.Value,
		RecID: r.ID,
		TTL:   ttl_t(r.TTL.Seconds()),
	}

	switch r.Type {
	case "URI":
		result.DTyp = "REDIRECT_URL"
	case "MX":
		result.Priority = ttl_t(r.Priority)
	case "HTTPS":
		tmp := make([]string, 0)
		tmp = append(tmp, strconv.Itoa(int(r.Priority)))
		tmp = append(tmp, r.Target)
		tmp = append(tmp, r.Value)
		result.DVal = strings.Join(tmp, " ")
	case "SRV":
		tmp := make([]string, 0)
		tmp = append(tmp, strconv.Itoa(int(r.Priority)))
		tmp = append(tmp, strconv.Itoa(int(r.Weight)))
		tmp = append(tmp, r.Value)
		result.DVal = strings.Join(tmp, " ")
	default:
	}
	return result
}

// AlidnsRecord convert libdns.Record with zone to aliDomaRecord
func alidnsRecordWithZone(r libdns.Record, zone string) aliDomaRecord {
	r.Name = libdns.RelativeName(r.Name, zone)
	rec := alidnsRecord(r)
	rec.DName = strings.Trim(zone, ".")
	return rec
}
