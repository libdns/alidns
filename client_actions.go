package alidns

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (c *aliClient) queryDomainInfo(ctx context.Context, zone string) (aliDomainInfo, error) {
	if c.schema == nil {
		return aliDomainInfo{}, errors.New("schema was not initialed proprely")
	}
	c.Lock()
	defer c.Unlock()
	c.AddRequestBody("Action", "DescribeDomains")
	c.AddRequestBody("KeyWord", zone)
	c.AddRequestBody("SearchMode", "EXACT")
	rs := aliDomainResult{}
	err := c.doAPIRequest(ctx, &rs)
	if err != nil {
		return aliDomainInfo{}, err
	}
	if len(rs.Domains.Domain) == 0 {
		return aliDomainInfo{}, errors.New("cannot found specified zone:" + zone)
	}
	return rs.Domains.Domain[0], err
}

func (c *aliClient) addDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	if c.schema == nil {
		return "", errors.New("schema was not initialed proprely")
	}
	c.Lock()
	defer c.Unlock()
	if rc.TTL <= 0 {
		rc.TTL = 600
	}
	c.AddRequestBody("Action", "AddDomainRecord")
	c.AddRequestBody("DomainName", rc.DomainName)
	c.AddRequestBody("RR", rc.Rr)
	c.AddRequestBody("Type", rc.DomainType)
	c.AddRequestBody("Value", rc.DomainValue)
	c.AddRequestBody("TTL", fmt.Sprintf("%d", rc.TTL))
	rs := aliDomainResult{}
	err = c.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (c *aliClient) delDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	if c.schema == nil {
		return "", errors.New("schema was not initialed proprely")
	}
	c.Lock()
	defer c.Unlock()
	c.AddRequestBody("Action", "DeleteDomainRecord")
	c.AddRequestBody("RecordId", rc.RecordID)
	rs := aliDomainResult{}
	err = c.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (c *aliClient) setDomainRecord(ctx context.Context, rc aliDomainRecord) (recID string, err error) {
	if c.schema == nil {
		return "", errors.New("schema was not initialed proprely")
	}
	c.Lock()
	defer c.Unlock()
	if rc.TTL <= 0 {
		rc.TTL = 600
	}
	c.AddRequestBody("Action", "UpdateDomainRecord")
	c.AddRequestBody("RecordId", rc.RecordID)
	c.AddRequestBody("RR", rc.Rr)
	c.AddRequestBody("Type", rc.DomainType)
	c.AddRequestBody("Value", rc.DomainValue)
	c.AddRequestBody("TTL", fmt.Sprintf("%d", rc.TTL))
	rs := aliDomainResult{}
	err = c.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (c *aliClient) getDomainRecord(ctx context.Context, recID string) (aliDomainRecord, error) {
	if c.schema == nil {
		return aliDomainRecord{}, errors.New("schema was not initialed proprely")
	}
	c.Lock()
	defer c.Unlock()
	c.AddRequestBody("Action", "DescribeDomainRecordInfo")
	c.AddRequestBody("RecordId", recID)
	rs := aliDomainResult{}
	err := c.doAPIRequest(ctx, &rs)
	rec := rs.ToDomaRecord()
	if err != nil {
		return aliDomainRecord{}, err
	}
	return rec, err
}

func (c *aliClient) queryDomainRecords(ctx context.Context, name string) ([]aliDomainRecord, error) {
	if c.schema == nil {
		return nil, errors.New("schema was not initialed proprely")
	}
	c.Lock()
	defer c.Unlock()
	c.AddRequestBody("Action", "DescribeDomainRecords")
	c.AddRequestBody("DomainName", strings.Trim(name, "."))
	rs := aliDomainResult{}
	err := c.doAPIRequest(ctx, &rs)
	if err != nil {
		return []aliDomainRecord{}, err
	}
	return rs.DomainRecords.Record, err
}

func (c *aliClient) queryDomainRecord(ctx context.Context, rr, name string, recType string, recVal ...string) (aliDomainRecord, error) {
	if c.schema == nil {
		return aliDomainRecord{}, errors.New("schema was not initialed proprely")
	}
	c.Lock()
	defer c.Unlock()
	c.AddRequestBody("Action", "DescribeDomainRecords")
	c.AddRequestBody("DomainName", strings.Trim(name, "."))
	c.AddRequestBody("RRKeyWord", rr)
	if recType != "" {
		c.AddRequestBody("TypeKeyWord", recType)
	}
	if len(recVal) > 0 && recVal[0] != "" {
		c.AddRequestBody("ValueKeyWord", recVal[0])
	}
	c.AddRequestBody("SearchMode", "ADVANCED")
	rs := aliDomainResult{}
	err := c.doAPIRequest(ctx, &rs)
	if err != nil {
		return aliDomainRecord{}, err
	}
	if len(rs.DomainRecords.Record) == 0 {
		return aliDomainRecord{}, errors.New("the Record Name of the domain not found")
	}
	return rs.DomainRecords.Record[0], err
}

// REVERSED:queryMainDomain rseserved for absolute names to name,zone
func (c *aliClient) queryMainDomain(ctx context.Context, name string) (string, string, error) {
	if c.schema == nil {
		return "", "", errors.New("schema was not initialed proprely")
	}
	c.Lock()
	defer c.Unlock()
	c.AddRequestBody("Action", "GetMainDomainName")
	c.AddRequestBody("InputString", strings.Trim(name, "."))
	rs := aliDomainResult{}
	err := c.doAPIRequest(ctx, &rs)
	if err != nil {
		return "", "", err
	}
	return rs.Rr, rs.DomainName, err
}
