package alidns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// aliClient is an abstration of AliClient
type aliClient struct {
	schema *aliClientSchema
	mutex  sync.Mutex
}

// TODO:Will complete,If we need to get Domain Info for something.
func (c *aliClient) getDomainInfo(ctx context.Context, zone string) error {
	return nil
}

func (c *aliClient) getClientSchema(cred *CredentialInfo, zone string) error {
	schema, err := getClientSchema(cred, "https")
	if err != nil {
		return err
	}
	c.schema = schema
	if zone != "" {
		c.getDomainInfo(context.Background(), strings.Trim(zone, "."))
	}
	return nil
}

func (c *aliClient) AddRequestBody(key string, value string) error {
	if c.schema == nil {
		return errors.New("schema was not initialed proprely")
	}
	return c.schema.addReqBody(key, value)
}

func (c *aliClient) SetRequestBody(key string, value string) error {
	if c.schema == nil {
		return errors.New("schema was not initialed proprely")
	}
	return c.schema.setReqBody(key, value)
}

func (c *aliClient) Lock() {
	c.mutex.Lock()
}

func (c *aliClient) Unlock() {
	c.mutex.Unlock()
}

func (c *aliClient) DoAPIRequest(ctx context.Context, method string, result interface{}) error {
	req, err := c.schema.HttpRequest(ctx, method)
	if err != nil {
		return err
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	var buf []byte
	buf, err = io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf, result)
	if err != nil {
		return err
	}
	if rsp.StatusCode != 200 {
		return fmt.Errorf("get error status: HTTP %d: %+v", rsp.StatusCode, result.(*aliDomainResult).Msg)
	}
	c.schema = nil
	return err
}
