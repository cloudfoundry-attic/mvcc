package mvcc

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	// By default, retry every 200ms for up to 1m
	DefaultDialRetries         = 100
	DefaultDialInterval        = 200 * time.Millisecond
	DefaultCloudControllerPath = "../../cloud_controller_ng"
)

type MVCC struct {
	cmd *exec.Cmd
}

func Dial(dialOptions ...DialOption) (*MVCC, error) {
	opts := &dialOpts{
		interval: DefaultDialInterval,
		retries:  DefaultDialRetries,
		ccPath:   DefaultCloudControllerPath,
	}
	for _, dialOption := range dialOptions {
		dialOption(opts)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(pwd, "cloud_controller.yml")

	if err := os.Chdir(opts.ccPath); err != nil {
		return nil, err
	}

	cmd := exec.Command(filepath.Join(opts.ccPath, "bin/cloud_controller"), "-c", configPath)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	mvcc := &MVCC{
		cmd: cmd,
	}

	if err := poll("http://localhost:8181/v2/info", opts.retries, opts.interval); err != nil {
		fmt.Fprintln(os.Stderr, mvcc.cmd.CombinedOutput())
		return nil, err
	}

	return mvcc, nil
}

func (mvcc *MVCC) Kill() error {
	return mvcc.cmd.Process.Kill()
}

type dialOpts struct {
	retries  int
	interval time.Duration
	ccPath   string
}

type DialOption func(*dialOpts)

func WithCloudControllerPath(path string) DialOption {
	return func(o *dialOpts) {
		o.ccPath = path
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
