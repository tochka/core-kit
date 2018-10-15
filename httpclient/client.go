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
	_, err := c.SendWithHeader(ctx, method, url, payload, nil, respObj)
	return err
}

func (c *VirgilHttpClient) SendWithHeader(ctx context.Context, method string, url string, payload interface{}, header http.Header, respObj interface{}) (http.Header, error) {
	var reqBody []byte
	var err error

	if payload != nil {
		switch v := payload.(type) {
		case []byte:
			reqBody = v
		case string:
			reqBody = []byte(v)
		default:
			reqBody, err = json.Marshal(payload)
			if err != nil {
				return nil, errors.Wrap(err, "HttpClient.Send [JSON marshal payload]")
			}
		}
	}
	req, err := http.NewRequest(method, fmt.Sprint(c.ServiceAddress, url), bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrapf(err, "HttpClient.Send [Method: %s Path: %s ]", method, url)
	}
	for k := range header {
		req.Header.Add(k, header.Get(k))
	}
	req.Header.Add("content-type", "application/json")

	resp, err := c.getHTTPClient().Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "HttpClient.Send [Send request]")
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotFound:
		return resp.Header, apierror.EntityNotFoundErr
	case http.StatusInternalServerError:
		return resp.Header, apierror.InternalServerErr
	case http.StatusUnauthorized:
		return resp.Header, apierror.UnauthorizedRequestErr
	case http.StatusForbidden:
		return resp.Header, apierror.ForbiddenErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.Header, errors.Wrapf(err, "HttpClient.Send [ReadBody (Method: %s Path: %s Body: %s)]", method, url, reqBody)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 { // http status code seccess
		var verr apierror.APIError
		if len(body) != 0 {
			err = json.Unmarshal(body, &verr)
			if err != nil {
				return resp.Header, errors.Wrapf(err, "HttpClient.Send [UnmarshalResponseErr(status code: %v body: %s)]", resp.StatusCode, body)
			}
		}
		verr.StatusCode = resp.StatusCode
		return resp.Header, verr
	}

	if respObj == nil {
		return resp.Header, nil
	}

	err = json.Unmarshal(body, respObj)
	if err != nil {
		return resp.Header, errors.Wrapf(err, "HttpClient.Send [UnmarshalResponseErr(status code: %v body: %s)]", resp.StatusCode, body)
	}
	return resp.Header, nil
}

func (c *VirgilHttpClient) getHTTPClient() HTTPClient {
	if c.Client == nil {
		return http.DefaultClient
	}
	return c.Client
}
