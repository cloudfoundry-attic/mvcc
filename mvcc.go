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

	"code.cloudfoundry.org/mvcc/internal/config"
	"github.com/phayes/freeport"
)

const (
	// By default, retry every 200ms for up to 1m
	DefaultDialRetries  = 100
	DefaultDialInterval = 200 * time.Millisecond

	DefaultHost = "localhost"
)

type MVCC struct {
	cmd    *exec.Cmd
	client *http.Client
	host   string
	port   int
}

func DialMVCC(dialOptions ...DialMVCCOption) (*MVCC, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	opts := &dialMVCCOpts{
		retries:  DefaultDialRetries,
		interval: DefaultDialInterval,
		configOptions: []config.Option{
			config.WithPort(port),
		},
	}
	for _, dialOption := range dialOptions {
		dialOption(opts)
	}

	ccConfigFile, err := config.Write(opts.configOptions...)
	if err != nil {
		return nil, err
	}
	defer ccConfigFile.Remove()

	ccBinaryPath, err := exec.LookPath("cloud_controller")
	if err != nil {
		return nil, ErrCCBinaryPathNotSet
	}

	cmd := exec.Command(ccBinaryPath, "-c", ccConfigFile.Name())
	cmd.Dir = filepath.Join(ccBinaryPath, "../..")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	cc := &MVCC{
		cmd:    cmd,
		client: &http.Client{},
		host:   DefaultHost,
		port:   port,
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

func (cc *MVCC) V2SetFeatureFlag(authToken string, flag string, enabled bool) error {
	body := v2FeatureFlagRequest{
		Enabled: enabled,
	}

	path := fmt.Sprintf("/v2/config/feature_flags/%s", flag)

	res, err := cc.Put(path, authToken, body, nil)
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		return convertStatusCode(200)
	}

	return nil
}

func (cc *MVCC) V3CreateOrganization(authToken string) (Organization, error) {
	var org Organization
	var o v3OrganizationResponse

	body := v3OrganizationRequest{
		Name: RandomUUID("org"),
	}

	res, err := cc.Post("/v3/organizations", authToken, body, &o)

	if err != nil {
		return org, err
	}
	if res.StatusCode == 201 {
		org.Name = o.Name
		org.UUID = o.GUID

		return org, nil
	}

	return org, convertStatusCode(res.StatusCode)
}

func (cc *MVCC) V3CreateSpace(authToken string, parentOrg Organization) (Space, error) {
	var space Space
	var s v3SpaceResponse

	var body v3SpaceRequest
	body.Name = RandomUUID("space")
	body.Relationships.Organization.Data.GUID = parentOrg.UUID

	res, err := cc.Post("/v3/spaces", authToken, body, &s)
	if err != nil {
		return space, err
	}
	if res.StatusCode != 201 {
		return space, convertStatusCode(res.StatusCode)
	}

	space.Name = s.Name
	space.UUID = s.GUID

	return space, nil
}

func (cc *MVCC) V3CreateApp(authToken string, parentSpace Space) (App, error) {
	var app App
	var a v3AppResponse

	var body v3AppRequest
	body.Name = RandomUUID("app")
	body.Relationships.Space.Data.GUID = parentSpace.UUID

	res, err := cc.Post("/v3/apps", authToken, body, &a)
	if err != nil {
		return app, err
	}
	if res.StatusCode != 201 {
		return app, convertStatusCode(res.StatusCode)
	}

	app.Name = a.Name
	app.UUID = a.GUID

	return app, nil
}

func (cc *MVCC) V3GetPackage(authToken string, uuid string) (Package, error) {
	var pkg Package
	var p v3PackageResponse

	path := fmt.Sprintf("/v3/packages/%s", uuid)

	res, err := cc.Get(path, authToken, &p)
	if err != nil {
		return pkg, err
	}
	if res.StatusCode != 200 {
		return pkg, convertStatusCode(res.StatusCode)
	}

	pkg.UUID = p.GUID
	pkg.Type = p.Type
	pkg.State = p.State

	return pkg, nil
}

func (cc *MVCC) V3CreatePackage(authToken string, parentApp App) (Package, error) {
	var pkg Package
	var p v3PackageResponse

	var body v3PackageRequest
	body.Relationships.App.Data.GUID = parentApp.UUID
	body.Data.Image = "alpine"
	body.Type = "docker"

	res, err := cc.Post("/v3/packages", authToken, body, &p)
	if err != nil {
		return pkg, err
	}
	if res.StatusCode != 201 {
		return pkg, convertStatusCode(res.StatusCode)
	}

	pkg.UUID = p.GUID
	pkg.Type = p.Type
	pkg.State = p.State

	return pkg, nil
}

func (cc *MVCC) V3GetBuild(authToken string, uuid string) (Build, error) {
	var b v3BuildResponse
	var build Build

	path := fmt.Sprintf("/v3/builds/%s", uuid)
	res, err := cc.Get(path, authToken, &b)
	if err != nil {
		return build, err
	}
	if res.StatusCode != 200 {
		return build, convertStatusCode(res.StatusCode)
	}

	build.UUID = b.GUID
	build.State = b.State
	build.DropletUUID = b.Droplet.GUID

	return build, nil
}

func (cc *MVCC) V3CreateBuild(authToken string, parentPackage Package) (Build, error) {
	var build Build
	var b v3BuildResponse

	var body v3BuildRequest
	body.Package.GUID = parentPackage.UUID

	res, err := cc.Post("/v3/builds", authToken, body, &b)
	if err != nil {
		return build, err
	}
	if res.StatusCode != 201 {
		var e V3ErrorResponse

		defer res.Body.Close()
		bits, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return build, err
		}

		if err = json.Unmarshal(bits, &e); err != nil {
			return build, err
		}

		fmt.Println("ERR:", e)

		return build, convertStatusCode(res.StatusCode)
	}

	build.UUID = b.GUID
	build.State = b.State
	build.DropletUUID = b.Droplet.GUID

	return build, nil
}

func (cc *MVCC) V3CreateTask(authToken string, parentApp App, dropletUUID string) (Task, error) {
	var task Task
	var t v3TaskResponse

	var body v3TaskRequest
	body.Command = "echo hello"
	body.DropletGUID = dropletUUID

	path := fmt.Sprintf("/v3/apps/%s/tasks", parentApp.UUID)
	res, err := cc.Post(path, authToken, body, &t)
	if err != nil {
		return task, err
	}
	if res.StatusCode != 202 {
		return task, convertStatusCode(res.StatusCode)
	}

	task.UUID = t.GUID

	return task, nil
}

func (cc *MVCC) V3GetTask(authToken string, taskUUID string) (Task, error) {
	var task Task
	var t v3TaskResponse

	path := fmt.Sprintf("/v3/tasks/%s", taskUUID)
	res, err := cc.Get(path, authToken, &t)
	if err != nil {
		return task, err
	}
	if res.StatusCode != 200 {
		return task, convertStatusCode(res.StatusCode)
	}

	task.UUID = t.GUID

	return task, nil
}

type dialMVCCOpts struct {
	retries  int
	interval time.Duration

	configOptions []config.Option
}

type DialMVCCOption func(*dialMVCCOpts)

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

func WithPermOptions(options PermOptions) DialMVCCOption {
	return func(o *dialMVCCOpts) {
		permOpts := []config.Option{
			config.WithPermEnabled(true),
			config.WithPermHostname("localhost"),
			config.WithPermPort(options.Port),
			config.WithPermCACertPath(options.CACertPath),
			config.WithPermTimeoutInMilliseconds(100),
		}

		o.configOptions = append(o.configOptions, permOpts...)
	}
}

func WithUAAOptions(options UAAOptions) DialMVCCOption {
	return func(o *dialMVCCOpts) {
		uaaURL := fmt.Sprintf("http://localhost:%d", options.Port)
		uaaOpts := []config.Option{
			config.WithUAAURL(uaaURL),
			config.WithUAAInternalURL(uaaURL),
		}

		o.configOptions = append(o.configOptions, uaaOpts...)
	}
}

type PermOptions struct {
	Port       int
	CACertPath string
}

type UAAOptions struct {
	Port int
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
