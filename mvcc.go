package mvcc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	// By default, retry every 200ms for up to 1m
	DefaultDialRetries  = 100
	DefaultDialInterval = 200 * time.Millisecond

	DefaultHost = "localhost"
	DefaultPort = 8181
)

type MVCC struct {
	cmd    *exec.Cmd
	client *http.Client
	host   string
	port   int
}

func DialMVCC(dialOptions ...DialMVCCOption) (*MVCC, error) {
	opts := &dialMVCCOpts{
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	cc := &MVCC{
		cmd:    cmd,
		client: &http.Client{},
		host:   DefaultHost,
		port:   DefaultPort,
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

func (cc *MVCC) Get(path string, authToken string, respData interface{}) (*http.Response, error) {
	return cc.Do("GET", path, authToken, nil, respData)
}

func (cc *MVCC) Post(path string, authToken string, body interface{}, respData interface{}) (*http.Response, error) {
	return cc.Do("POST", path, authToken, body, respData)
}

func (cc *MVCC) Put(path string, authToken string, body interface{}, respData interface{}) (*http.Response, error) {
	return cc.Do("PUT", path, authToken, body, respData)
}

func (cc *MVCC) Delete(path string, authToken string) (*http.Response, error) {
	return cc.Do("DELETE", path, authToken, nil, nil)
}

func (cc *MVCC) Do(verb string, path string, authToken string, body interface{}, respData interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		bodyBits, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(bodyBits)
	}

	req, err := http.NewRequest(verb, fmt.Sprintf("http://%s:%d%s", cc.host, cc.port, path), reqBody)
	if err != nil {
		return nil, err
	}

	if verb == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	res, err := cc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bits, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if respData != nil && res.StatusCode >= 200 && res.StatusCode < 300 {
		err = json.Unmarshal(bits, respData)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (cc *MVCC) V3CreateOrganization(authToken string) (Organization, error) {
	var org Organization
	u, err := RandomUUID("org")
	if err != nil {
		return org, err
	}
	body := V3OrganizationRequest{
		Name: u,
	}

	res, err := cc.Post("/v3/organizations", authToken, body, &org)

	if err != nil {
		return org, err
	}
	if res.StatusCode == 201 {
		return org, nil
	}
	return org, convertStatusCode(res.StatusCode)
}

type dialMVCCOpts struct {
	retries      int
	interval     time.Duration
	ccPath       string
	ccConfigPath string
}

type DialMVCCOption func(*dialMVCCOpts)

func WithCloudControllerPath(path string) DialMVCCOption {
	return func(o *dialMVCCOpts) {
		o.ccPath = path
	}
}

func WithCloudControllerConfigPath(path string) DialMVCCOption {
	return func(o *dialMVCCOpts) {
		o.ccConfigPath = path
	}
}

func WithDialRetries(retries int) DialMVCCOption {
	return func(o *dialMVCCOpts) {
		o.retries = retries
	}
}

func WithDialRetryInterval(interval time.Duration) DialMVCCOption {
	return func(o *dialMVCCOpts) {
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

func convertStatusCode(statusCode int) error {
	switch statusCode {
	case 400:
		return ErrBadRequest
	case 401:
		return ErrUnauthenticated
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	default:
		return &ErrUnexpectedStatusCode{
			StatusCode: statusCode,
		}
	}
}
