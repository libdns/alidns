package alidns

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for Alicloud.
type Provider struct {
	client aliClient
	// The API Key ID Required by Aliyun's for accessing the Aliyun's API
	AccKeyID string `json:"access_key_id"`
	// The API Key Secret Required by Aliyun's for accessing the Aliyun's API
	AccKeySecret string `json:"access_key_secret"`
	// Optional for identifing the region of the Aliyun's Service,The default is zh-hangzhou
	RegionID string `json:"region_id,omitempty"`
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	var errs = OpErrors("AppendRecords")
	for _, rec := range recs {
		ar := alidnsRecord(rec, zone)
		rid, err := p.addDomainRecord(ctx, ar)
		if err != nil {
			errs.JoinRecord(rec, err)
		}
		ar.RecordID = rid
		rls = append(rls, ar)
	}
	return rls, errs.Error()
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	var errs = OpErrors("DeleteRecords")
	for _, rec := range recs {
		ar := alidnsRecord(rec, zone)
		if ar.RecordID == "" {
			r0, err := p.queryDomainRecord(ctx, ar.Rr, ar.DomainName, ar.DomainType, ar.DomainValue)
			if err != nil {
				errs.JoinRecord(rec, err)
			}
			ar.RecordID = r0.RecordID
		}
		_, err := p.delDomainRecord(ctx, ar)
		if err != nil {
			errs.JoinRecord(rec, err)
		}
		rls = append(rls, ar)
	}
	return rls, errs.Error()
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	var rls []libdns.Record
	recs, err := p.queryDomainRecords(ctx, zone)
	if err != nil {
		return nil, OpError("GetRecords", err)
	}
	for _, rec := range recs {
		rls = append(rls, rec)
	}
	return rls, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	var errs = OpErrors("SetRecords")
	for _, rec := range recs {
		ar := alidnsRecord(rec, zone)
		if ar.RecordID == "" {
			r0, err := p.queryDomainRecord(ctx, ar.Rr, ar.DomainName, ar.DomainType, ar.DomainValue)
			if err != nil {
				ar.RecordID, err = p.addDomainRecord(ctx, ar)
				if err != nil {
					errs.JoinRecord(rec, err)
				}
			} else {
				ar.RecordID = r0.RecordID
			}
		} else {
			_, err := p.setDomainRecord(ctx, ar)
			if err != nil {
				errs.JoinRecord(rec, err)
			}
		}
		rls = append(rls, ar)
	}
	return rls, errs.Error()
}

func (p *Provider) getClient() error {
	return p.getClientWithZone("")
}

func (p *Provider) getClientWithZone(zone string) error {
	cred := NewCredentialInfo(p.AccKeyID, p.AccKeySecret, p.RegionID)
	return p.client.getClientSchema(cred, zone)
}

func (p *Provider) addDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	p.client.Lock()
	defer p.client.Unlock()
	p.getClientWithZone(rc.DomainName)
	if rc.TTL <= 0 {
		rc.TTL = 600
	}
	p.client.AddRequestBody("Action", "AddDomainRecord")
	p.client.AddRequestBody("DomainName", rc.DomainName)
	p.client.AddRequestBody("RR", rc.Rr)
	p.client.AddRequestBody("Type", rc.DomainType)
	p.client.AddRequestBody("Value", rc.DomainValue)
	p.client.AddRequestBody("TTL", fmt.Sprintf("%d", rc.TTL))
	rs := aliDomainResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) delDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	p.client.Lock()
	defer p.client.Unlock()
	p.getClient()
	p.client.AddRequestBody("Action", "DeleteDomainRecord")
	p.client.AddRequestBody("RecordId", rc.RecordID)
	rs := aliDomainResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) setDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	p.client.Lock()
	defer p.client.Unlock()
	p.getClientWithZone(rc.DomainName)
	if rc.TTL <= 0 {
		rc.TTL = 600
	}
	p.client.AddRequestBody("Action", "UpdateDomainRecord")
	p.client.AddRequestBody("RecordId", rc.RecordID)
	p.client.AddRequestBody("RR", rc.Rr)
	p.client.AddRequestBody("Type", rc.DomainType)
	p.client.AddRequestBody("Value", rc.DomainValue)
	p.client.AddRequestBody("TTL", fmt.Sprintf("%d", rc.TTL))
	rs := aliDomainResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) getDomainRecord(ctx context.Context, recID string) (aliDomainRecord, error) {
	p.client.Lock()
	defer p.client.Unlock()
	p.getClient()
	p.client.AddRequestBody("Action", "DescribeDomainRecordInfo")
	p.client.AddRequestBody("RecordId", recID)
	rs := aliDomainResult{}
	err := p.doAPIRequest(ctx, &rs)
	rec := rs.ToDomaRecord()
	if err != nil {
		return aliDomainRecord{}, err
	}
	return rec, err
}

func (p *Provider) queryDomainRecords(ctx context.Context, name string) ([]aliDomainRecord, error) {
	p.client.Lock()
	defer p.client.Unlock()
	p.getClient()
	p.client.AddRequestBody("Action", "DescribeDomainRecords")
	p.client.AddRequestBody("DomainName", strings.Trim(name, "."))
	rs := aliDomainResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return []aliDomainRecord{}, err
	}
	return rs.DomainRecords.Record, err
}

func (p *Provider) queryDomainRecord(ctx context.Context, rr, name string, recType string, recVal ...string) (aliDomainRecord, error) {
	p.client.Lock()
	defer p.client.Unlock()
	p.getClient()
	p.client.AddRequestBody("Action", "DescribeDomainRecords")
	p.client.AddRequestBody("DomainName", strings.Trim(name, "."))
	p.client.AddRequestBody("RRKeyWord", rr)
	if recType != "" {
		p.client.AddRequestBody("TypeKeyWord", recType)
	}
	if len(recVal) > 0 && recVal[0] != "" {
		p.client.AddRequestBody("ValueKeyWord", recVal[0])
	}
	p.client.AddRequestBody("SearchMode", "ADVANCED")
	rs := aliDomainResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return aliDomainRecord{}, err
	}
	if len(rs.DomainRecords.Record) == 0 {
		return aliDomainRecord{}, errors.New("the Record Name of the domain not found")
	}
	return rs.DomainRecords.Record[0], err
}

// REVERSED:queryMainDomain rseserved for absolute names to name,zone
func (p *Provider) queryMainDomain(ctx context.Context, name string) (string, string, error) {
	p.client.Lock()
	defer p.client.Unlock()
	p.getClient()
	p.client.AddRequestBody("Action", "GetMainDomainName")
	p.client.AddRequestBody("InputString", strings.Trim(name, "."))
	rs := aliDomainResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return "", "", err
	}
	return rs.Rr, rs.DomainName, err
}

func (p *Provider) doAPIRequest(ctx context.Context, result interface{}) error {
	return p.client.DoAPIRequest(ctx, "GET", result)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
