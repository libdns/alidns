package alidns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

// Client is an abstration of AliClient
type Client struct {
	AClient *AliClient
	mutex   sync.Mutex
}

func (p *Provider) getClient() error {
	return p.getClientWithZone("")
}

func (p *Provider) getClientWithZone(zone string) error {
	cred := newCredInfo(p.AccKeyID, p.AccKeySecret, p.RegionID)
	return p.client.getAliClient(cred, zone)
}

func (p *Provider) addDomainRecord(ctx context.Context, rc aliDomaRecord) (recID string, err error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClientWithZone(rc.DName)
	p.client.AClient.addReqBody("Action", "AddDomainRecord")
	p.client.AClient.addReqBody("DomainName", rc.DName)
	p.client.AClient.addReqBody("RR", rc.Rr)
	p.client.AClient.addReqBody("Type", rc.DTyp)
	p.client.AClient.addReqBody("Value", rc.DVal)
	p.client.AClient.addReqBody("TTL", fmt.Sprintf("%d", rc.TTL))
	rs := aliResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) delDomainRecord(ctx context.Context, rc aliDomaRecord) (recID string, err error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.AClient.addReqBody("Action", "DeleteDomainRecord")
	p.client.AClient.addReqBody("RecordId", rc.RecID)
	rs := aliResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) setDomainRecord(ctx context.Context, rc aliDomaRecord) (recID string, err error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClientWithZone(rc.DName)
	p.client.AClient.addReqBody("Action", "UpdateDomainRecord")
	p.client.AClient.addReqBody("RecordId", rc.RecID)
	p.client.AClient.addReqBody("RR", rc.Rr)
	p.client.AClient.addReqBody("Type", rc.DTyp)
	p.client.AClient.addReqBody("Value", rc.DVal)
	p.client.AClient.addReqBody("TTL", fmt.Sprintf("%d", rc.TTL))
	rs := aliResult{}
	err = p.doAPIRequest(ctx, &rs)
	recID = rs.RecID
	if err != nil {
		return "", err
	}
	return recID, err
}

func (p *Provider) getDomainRecord(ctx context.Context, recID string) (aliDomaRecord, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.AClient.addReqBody("Action", "DescribeDomainRecordInfo")
	p.client.AClient.addReqBody("RecordId", recID)
	rs := aliResult{}
	err := p.doAPIRequest(ctx, &rs)
	rec := rs.ToDomaRecord()
	if err != nil {
		return aliDomaRecord{}, err
	}
	return rec, err
}

func (p *Provider) queryDomainRecords(ctx context.Context, name string) ([]aliDomaRecord, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.AClient.addReqBody("Action", "DescribeDomainRecords")
	p.client.AClient.addReqBody("DomainName", name)
	rs := aliResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return []aliDomaRecord{}, err
	}
	return rs.DRecords.Record, err
}

func (p *Provider) queryDomainRecord(ctx context.Context, rr string, name string) (aliDomaRecord, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.AClient.addReqBody("Action", "DescribeDomainRecords")
	p.client.AClient.addReqBody("DomainName", name)
	p.client.AClient.addReqBody("KeyWord", rr)
	p.client.AClient.addReqBody("SearchMode", "EXACT")
	rs := aliResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return aliDomaRecord{}, err
	}
	if len(rs.DRecords.Record) == 0 {
		return aliDomaRecord{}, errors.New("the Record Name of the domain not found")
	}
	return rs.DRecords.Record[0], err
}

// REVERSED:queryMainDomain rseserved for absolute names to name,zone
func (p *Provider) queryMainDomain(ctx context.Context, name string) (string, string, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()
	p.getClient()
	p.client.AClient.addReqBody("Action", "GetMainDomainName")
	p.client.AClient.addReqBody("InputString", name)
	rs := aliResult{}
	err := p.doAPIRequest(ctx, &rs)
	if err != nil {
		return "", "", err
	}
	return rs.Rr, rs.DName, err
}

func (p *Provider) doAPIRequest(ctx context.Context, result interface{}) error {
	return p.client.doAPIRequest(ctx, "GET", &result)
}

// TODO:Will complete,If we need to get Domain Info for something.
func (c *Client) getDomainInfo(ctx context.Context, zone string) error {
	return nil
}

func (c *Client) doAPIRequest(ctx context.Context, method string, result interface{}) error {
	req, err := c.applyReq(ctx, method, nil)
	if err != nil {
		return err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	var buf []byte
	buf, err = ioutil.ReadAll(rsp.Body)
	strBody := string(buf)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(strBody), result)
	if err != nil {
		return err
	}
	if rsp.StatusCode != 200 {
		return fmt.Errorf("get error status: HTTP %d: %+v", rsp.StatusCode, result.(*aliResult).Msg)
	}
	c.AClient = nil
	return err
}
