package mvcc

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type PermServer struct {
	cmd *exec.Cmd
}

func DialPermServer(dialOptions ...DialPermServerOption) (*PermServer, error) {
	opts := &dialPermServerOpts{}
	for _, dialOption := range dialOptions {
		dialOption(opts)
	}

	if opts.permServerBinaryPath == "" {
		return nil, ErrPermServerBinaryPathNotSet
	}
	if opts.permServerCertsPath == "" {
		return nil, ErrPermServerCertsPathNotSet
	}

	cmd := exec.Command(
		filepath.Join(opts.permServerBinaryPath, "perm"),
		"serve",
		"--tls-certificate", filepath.Join(opts.permServerCertsPath, "perm-server.crt"),
		"--tls-key", filepath.Join(opts.permServerCertsPath, "perm-server.key"),
		"--log-level", "debug",
		"--listen-port", "3333",
		"--db-driver", "mysql",
		"--db-host", "localhost",
		"--db-port", "3306",
		"--db-username", "root",
		"--db-password", "password",
		"--db-schema", "perm",
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// time for the server to energize
	time.Sleep(100 * time.Millisecond)

	return &PermServer{
		cmd: cmd,
	}, nil
}

func (p *PermServer) Kill() error {
	return p.cmd.Process.Kill()
}

type dialPermServerOpts struct {
	permServerBinaryPath string
	permServerCertsPath  string
}

type DialPermServerOption func(*dialPermServerOpts)

func WithPermBinaryPath(path string) DialPermServerOption {
	return func(o *dialPermServerOpts) {
		o.permServerBinaryPath = path
	}
}

func WithPermCertsPath(path string) DialPermServerOption {
	return func(o *dialPermServerOpts) {
		o.permServerCertsPath = path
	}
}
