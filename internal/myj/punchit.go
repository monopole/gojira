package myj

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/monopole/gojira/internal/myhttp"
	"github.com/monopole/gojira/internal/utils"
	"io"
	"net/http"
	"net/url"
)

func (jb *JiraBoss) punchItChewie(
	method string, req any, path string) ([]byte, error) {
	loc, err := url.Parse(myhttp.Scheme + jb.args.Host + "/" + path)
	if err != nil {
		return nil, err
	}
	var (
		ans  io.ReadCloser
		body []byte
	)
	body, err = json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf(
			"trouble marshaling data from %s request; %w", method, err)
	}
	if req != nil && utils.Debug {
		dump("REQUEST", body)
	}
	ans, err = jb.doRequest(method, loc, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	body, err = io.ReadAll(ans)
	if err != nil {
		return nil, fmt.Errorf("ReadAll failure: %w", err)
	}
	if utils.Debug {
		dump("RESPONSE", body)
	}
	return body, nil
}

func dump(lab string, body []byte) {
	utils.DoErrF("BEGIN %s ---------------------", lab)
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, " ", "  "); err != nil {
		utils.DoErr1(fmt.Errorf("json indent failure %w", err).Error())
	} else {
		utils.DoErr1(pretty.String())
	}
	utils.DoErrF("END %s ---------------------", lab)
}

func (jb *JiraBoss) doRequest(
	method string, loc *url.URL,
	reqBody io.Reader) (ans io.ReadCloser, err error) {
	var (
		req  *http.Request
		resp *http.Response
	)
	req, err = http.NewRequest(method, loc.String(), reqBody)
	if err != nil {
		return
	}
	req.Header.Set(myhttp.HeaderAccept, myhttp.ContentTypeJson)
	req.Header.Set(myhttp.HeaderContentType, myhttp.ContentTypeJson)
	req.Header.Set(myhttp.HeaderAAuthorization,
		fmt.Sprintf("Bearer: %s", jb.args.Token))
	if utils.Debug {
		_ = myhttp.PrintRequest(req, myhttp.PrArgs{Headers: true, Body: false})
	}
	resp, err = jb.htCl.Do(req)
	if err != nil {
		if utils.Debug && resp != nil {
			_ = myhttp.PrintResponse(resp, myhttp.PrArgs{Headers: true, Body: true})
		}
		return
	}
	if !(resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusNoContent ||
		resp.StatusCode == http.StatusAccepted ||
		resp.StatusCode == http.StatusCreated) {
		if utils.Debug {
			_ = myhttp.PrintResponse(resp, myhttp.PrArgs{Headers: true, Body: true})
		}
		err = fmt.Errorf("status code %d", resp.StatusCode)
		return
	}
	return resp.Body, nil
}
