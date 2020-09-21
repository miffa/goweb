package tpaashttpclient

import (
	"crypto/tls"
	"encoding/json"

	"github.com/parnurzeal/gorequest"

	"github.com/pkg/errors"
)

const ()

type HttpClient struct {
	U      string
	P      string
	Domain string
}

func NewDefaultHttpClient() *HttpClient {
	return &HttpClient{U: "",
		P:      "",
		Domain: ""}
}

func NewHttpClient(u, p, domain string) *HttpClient {
	return &HttpClient{U: u,
		P:      p,
		Domain: domain}
}

func (h *HttpClient) Req() *gorequest.SuperAgent {
	request := gorequest.New().
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	if h.U == "" || h.P == "" {
		return request
	}

	request.SetBasicAuth(h.U, h.P)
	return request
}

/*
 making "/search?query=bicycle&size=50x50&weight=20kg"
gorequest.New().
  Get("/search").
  Query(`{ query: 'bicycle' }`).
  Query(`{ size: '50x50' }`).
  Query(`{ weight: '20kg' }`).
  End()
or
gorequest.New().
  Get("/search").
  Query(`{ query: 'bicycle', size: '50x50', weight: '20kg' }`).
  End()
*/

func (h *HttpClient) Get(url string, query interface{}, data interface{}) error {
	be := Begin()
	defer End(url, be)
	resp, body, errs := h.Req().Get(url).Query(query).End()
	if len(errs) != 0 {
		return errs[0]
	}

	if HttpErr(resp.StatusCode) {
		return errors.Errorf("%s.%s", resp.Status, body)
	}

	if data != nil {
		json.Unmarshal([]byte(body), data)
	}
	return nil
}

func (h *HttpClient) Head(url string, query interface{}, data interface{}) error {
	be := Begin()
	defer End(url, be)
	resp, body, errs := h.Req().Head(url).Query(query).End()
	if len(errs) != 0 {
		return errs[0]
	}

	if HttpErr(resp.StatusCode) {
		return errors.Errorf("%s.%s", resp.Status, body)
	}

	if data != nil {
		json.Unmarshal([]byte(body), data)
	}
	return nil
}

func (h *HttpClient) PostForm(url string, data interface{}, ret interface{}) error {
	be := Begin()
	defer End(url, be)
	resp, body, errs := h.Req().Post(url).Send(data).End()
	if len(errs) != 0 {
		return errs[0]
	}

	if HttpErr(resp.StatusCode) {
		return errors.Errorf("%s.%s", resp.Status, body)
	}

	if ret != nil {
		json.Unmarshal([]byte(body), ret)
	}
	return nil
}

func (h *HttpClient) PostJson(url string, data interface{}, ret interface{}) error {
	be := Begin()
	defer End(url, be)
	resp, body, errs := h.Req().Post(url).Send(data).Type("json").End()
	if len(errs) != 0 {
		return errs[0]
	}

	if HttpErr(resp.StatusCode) {
		return errors.Errorf("%s.%s", resp.Status, body)
	}

	if ret != nil {
		json.Unmarshal([]byte(body), ret)
	}
	return nil
}

func (h *HttpClient) PutForm(url string, data interface{}, ret interface{}) error {
	be := Begin()
	defer End(url, be)
	resp, body, errs := h.Req().Put(url).Send(data).End()
	if len(errs) != 0 {
		return errs[0]
	}

	if HttpErr(resp.StatusCode) {
		return errors.Errorf("%s.%s", resp.Status, body)
	}

	if ret != nil {
		json.Unmarshal([]byte(body), ret)
	}
	return nil
}

func (h *HttpClient) PutJson(url string, data interface{}, ret interface{}) error {
	be := Begin()
	defer End(url, be)
	resp, body, errs := h.Req().Post(url).Send(data).Type("json").End()
	if len(errs) != 0 {
		return errs[0]
	}

	if HttpErr(resp.StatusCode) {
		return errors.Errorf("%s.%s", resp.Status, body)
	}

	if ret != nil {
		json.Unmarshal([]byte(body), ret)
	}
	return nil
}

func (h *HttpClient) Delete(url string) error {
	be := Begin()
	defer End(url, be)
	resp, body, errs := h.Req().Delete(url).End()
	if len(errs) != 0 {
		return errs[0]
	}

	if Http200(resp.StatusCode) {
		return nil
	}

	return errors.Errorf("%s.%s", resp.Status, body)
}
