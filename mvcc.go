package mvcc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	// By default, retry every 200ms for up to 1m
	DefaultDialRetries  = 100
	DefaultDialInterval = 200 * time.Millisecond
)

var validToken string = `bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwidHlwIjoiSldUIn0.eyJqdGkiOiI5YmUxODkyYzcyYTM0NzJkOGY4MGQxMWZjOTgyNTc4NCIsInN1YiI6IjRkM2UwNGIxLWY4OWYtNDM3MC1hZGE3LTcwZThkMWI3ZjNjMSIsInNjb3BlIjpbImNsb3VkX2NvbnRyb2xsZXIucmVhZCIsInBhc3N3b3JkLndyaXRlIiwiY2xvdWRfY29udHJvbGxlci53cml0ZSIsIm9wZW5pZCIsInVhYS51c2VyIl0sImNsaWVudF9pZCI6ImNmIiwiY2lkIjoiY2YiLCJhenAiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI0ZDNlMDRiMS1mODlmLTQzNzAtYWRhNy03MGU4ZDFiN2YzYzEiLCJvcmlnaW4iOiJ1YWEiLCJ1c2VyX25hbWUiOiJmb28iLCJlbWFpbCI6ImZvbyIsInJldl9zaWciOiI2NjZhNjUxMCIsImlhdCI6MTUyOTUyMjg5MywiZXhwIjoxNTI5NTIzNDkzLCJpc3MiOiJodHRwczovL3VhYS5mb3NzaWwtbXVzdGFuZy5jYXBpLmxhbmQvb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIiwiY2YiLCJ1YWEiLCJvcGVuaWQiXX0.nEcML5IifdW_CIonUM-ebY1RRspJlsr7fVq6pHPQLTnOhMmi2ZR2lDzvbQ99LTS-kd0E2juzuOLYQYryPutbyLm2LgwbtvCRD9IxNnwGYPwIVlodfHdCqocMQXEtvlSdNGfY1kwAGtv9NJowPDVpxJKE4H1Hxx0MRObFnJcH_W9F4yJAUf5ALjplFOzsmmnfsqTDfTGR2oo24133YYCGSyOaUBiOghJvUG7IQiFdVjR_7yxSObj00DO6VbWlEGTYLbChuN37Hm90Fu9cCllheSy0pElfdLKt6lojcSGG1pMnwpnK74n3n3qoBlK3poVsosj6g_xfkuLfQ_Y4yiq1IQ`

type MVCC struct {
	cmd    *exec.Cmd
	client *http.Client
	host   string
	port   int
}

func Dial(dialOptions ...DialOption) (*MVCC, error) {
	opts := &dialOpts{
		retries:  DefaultDialRetries,
		interval: DefaultDialInterval,
	}
	for _, dialOption := range dialOptions {
		dialOption(opts)
	}

	if opts.ccPath == "" {
		return nil, ErrCCBinaryPathNotSet
	}
	if opts.ccConfigPath == "" {
		return nil, ErrCCConfigPathNotSet
	}

	if err := os.Chdir(opts.ccPath); err != nil {
		return nil, err
	}

	cmd := exec.Command(
		filepath.Join(opts.ccPath, "bin/cloud_controller"),
		"-c", opts.ccConfigPath,
	)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	cc := &MVCC{
		cmd:    cmd,
		client: &http.Client{},
		host:   "localhost",
		port:   8181,
	}

	if err := poll(fmt.Sprintf("http://%s:%d/v2/info", cc.host, cc.port), opts.retries, opts.interval); err != nil {
		out, cErr := cc.cmd.CombinedOutput()
		if cErr != nil {
			fmt.Fprintf(os.Stderr, "getting combined output error: %s\n", cErr.Error())
		} else {
			fmt.Fprintf(os.Stderr, "combined output: %s\n", out)
		}
		return nil, err
	}

	return cc, nil
}

func (cc *MVCC) Kill() error {
	return cc.cmd.Process.Kill()
}

func (cc *MVCC) Get(path string, respData interface{}) (*http.Response, error) {
	res, err := cc.client.Do(&http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", cc.host, cc.port),
			Path:   path,
		},
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bits, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bits, respData)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cc *MVCC) Post(path string, body interface{}, respData interface{}) (*http.Response, error) {
	bodyBits, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", validToken)

	res, err := cc.client.Do(&http.Request{
		Body:   ioutil.NopCloser(bytes.NewBuffer(bodyBits)),
		Header: headers,
		Method: "POST",
		URL: &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", cc.host, cc.port),
			Path:   path,
		},
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bits, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s\n", bits)

	err = json.Unmarshal(bits, respData)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type dialOpts struct {
	retries      int
	interval     time.Duration
	ccPath       string
	ccConfigPath string
}

type DialOption func(*dialOpts)

func WithCloudControllerPath(path string) DialOption {
	return func(o *dialOpts) {
		o.ccPath = path
	}
}

func WithCloudControllerConfigPath(path string) DialOption {
	return func(o *dialOpts) {
		o.ccConfigPath = path
	}
}

func WithDialRetries(retries int) DialOption {
	return func(o *dialOpts) {
		o.retries = retries
	}
}

func WithDialRetryInterval(interval time.Duration) DialOption {
	return func(o *dialOpts) {
		o.interval = interval
	}
}

func poll(addr string, retries int, interval time.Duration) error {
	for i := 0; i < retries; i++ {
		resp, err := http.Get(addr)
		if err == nil && resp.StatusCode == 200 {
			return nil
		}

		time.Sleep(interval)
	}

	return ErrFailedToStart
}
