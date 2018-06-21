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

func (cc *MVCC) Get(path string, authToken string, respData interface{}) (*http.Response, error) {
	headers := http.Header{}
	if authToken != "" {
		headers.Set("Authorization", authToken)
	}

	res, err := cc.client.Do(&http.Request{
		Header: headers,
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

func (cc *MVCC) Post(path string, authToken string, bodyBits []byte, respData interface{}) (*http.Response, error) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	if authToken != "" {
		headers.Set("Authorization", authToken)
	}

	res, err := cc.client.Do(&http.Request{
		Body:   ioutil.NopCloser(bytes.NewReader(bodyBits)),
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

	if res.StatusCode == 200 {
		err = json.Unmarshal(bits, respData)
		if err != nil {
			return nil, err
		}
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
