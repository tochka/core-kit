package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tochka/core-kit/apierror"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type VirgilHttpClient struct {
	Client         HTTPClient
	ServiceAddress string
}

func (c *VirgilHttpClient) Send(ctx context.Context, method string, url string, payload interface{}, respObj interface{}) error {
	var reqBody []byte
	var err error

	if payload != nil {
		reqBody, err = json.Marshal(payload)
		if err != nil {
			return errors.Wrap(err, "HttpClient.Send [JSON marshal payload]")
		}
	}
	req, err := http.NewRequest(method, fmt.Sprint(c.ServiceAddress, url), bytes.NewReader(reqBody))
	if err != nil {
		return errors.Wrapf(err, "HttpClient.Send [Method: %s Path: %s ]", method, url)
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.getHTTPClient().Do(req)
	if err != nil {
		return errors.Wrapf(err, "HttpClient.Send [Send request]")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return apierror.EntityNotFoundErr
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "HttpClient.Send [ReadBody (Method: %s Path: %s Body: %s)]", method, url, reqBody)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 { // http status code seccess
		var verr apierror.APIError
		err = json.Unmarshal(body, &verr)
		if err != nil {
			return errors.Wrapf(err, "HttpClient.Send [UnmarshalResponseErr(status code: %v body: %s)]", resp.StatusCode, body)
		}
		verr.StatusCode = resp.StatusCode
		return verr
	}

	if respObj == nil {
		return nil
	}

	err = json.Unmarshal(body, respObj)
	if err != nil {
		return errors.Wrapf(err, "HttpClient.Send [UnmarshalResponseErr(status code: %v body: %s)]", resp.StatusCode, body)
	}
	return nil
}

func (c *VirgilHttpClient) getHTTPClient() HTTPClient {
	if c.Client == nil {
		return http.DefaultClient
	}
	return c.Client
}
