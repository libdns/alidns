package alidns

import (
	"context"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for Alicloud.
type Provider struct {
	client *aliClient
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
			continue
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
				continue
			}
			ar.RecordID = r0.RecordID
		}
		_, err := p.delDomainRecord(ctx, ar)
		if err != nil {
			errs.JoinRecord(rec, err)
			continue
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
			r0, err := p.client.queryDomainRecord(ctx, ar.Rr, ar.DomainName, ar.DomainType, ar.DomainValue)
			if err != nil {
				ar.RecordID, err = p.client.addDomainRecord(ctx, ar)
				if err != nil {
					errs.JoinRecord(rec, err)
					continue
				}
			} else {
				ar.RecordID = r0.RecordID
			}
		} else {
			_, err := p.setDomainRecord(ctx, ar)
			if err != nil {
				errs.JoinRecord(rec, err)
				continue
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
	var err error
	p.client, err = getClient(cred, zone)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) addDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	err = p.getClientWithZone(rc.DomainName)
	if err != nil {
		return "", err
	}
	return p.client.addDomainRecord(ctx, rc)
}

func (p *Provider) delDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	err = p.getClientWithZone(rc.DomainName)
	if err != nil {
		return "", err
	}
	return p.client.delDomainRecord(ctx, rc)
}

func (p *Provider) setDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	err = p.getClientWithZone(rc.DomainName)
	if err != nil {
		return "", err
	}
	return p.client.setDomainRecord(ctx, rc)
}

func (p *Provider) getDomainRecord(ctx context.Context, recID string) (aliDomainRecord, error) {
	p.getClient()
	return p.client.getDomainRecord(ctx, recID)
}

func (p *Provider) queryDomainRecords(ctx context.Context, name string) ([]aliDomainRecord, error) {
	p.getClient()
	return p.client.queryDomainRecords(ctx, name)
}

func (p *Provider) queryDomainRecord(ctx context.Context, rr, name string, recType string, recVal ...string) (aliDomainRecord, error) {
	p.getClient()
	return p.client.queryDomainRecord(ctx, rr, name, recType, recVal...)
}

// REVERSED:queryMainDomain rseserved for absolute names to name,zone
func (p *Provider) queryMainDomain(ctx context.Context, name string) (string, string, error) {
	p.getClient()
	return p.client.queryMainDomain(ctx, name)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
