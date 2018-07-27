package diegox

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/mvcc"
)

type BBSServer struct {
	logger lager.Logger
	mux    *http.ServeMux
	server *http.Server
}

func NewBBSServer(opts ...BBSServerOption) *BBSServer {
	o := defaultBBSServerOptions()
	for _, opt := range opts {
		opt(o)
	}

	logger := o.logger

	mux := &http.ServeMux{}
	mux.HandleFunc("/v1/tasks/desire.r2", desireTaskHandler(logger))

	return &BBSServer{
		logger: logger,
		mux:    mux,
	}
}

func (s *BBSServer) ListenAndServe(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.Serve(listener)
}

func (s *BBSServer) Serve(listener net.Listener) error {
	if s.server == nil {
		s.server = &http.Server{
			Handler: s.mux,
		}
	}

	return s.server.Serve(listener)
}

type BBSServerOption func(*bbsServerOptions)

func WithLogger(logger lager.Logger) BBSServerOption {
	return func(o *bbsServerOptions) {
		o.logger = logger
	}
}

type bbsServerOptions struct {
	logger lager.Logger
}

func defaultBBSServerOptions() *bbsServerOptions {
	return &bbsServerOptions{
		logger: lagertest.NewTestLogger("fake-bbs"),
	}
}

func desireTaskHandler(logger lager.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("started /v1/tasks/desire.r2")
		defer logger.Debug("finished /v1/tasks/desire.r2")

		defer r.Body.Close()

		bits, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			logger.Error("failed to read body", err)
			return
		}

		req := &models.DesireTaskRequest{}
		if err = req.Unmarshal(bits); err != nil {
			w.WriteHeader(500)
			logger.Error("failed to unmarshal DesireTaskRequest", err)
			return
		}

		body := &taskCallbackRequest{}
		body.Result.LifecycleMetadata.DockerImage = "alpine"
		body.Result.LifecycleType = string(mvcc.DockerType)
		body.Result.ProcessTypes = map[string]string{
			"docker": "docker run",
		}
		body.TaskGUID = req.TaskGuid
		body.Result.TaskGUID = req.TaskGuid

		b, err := json.Marshal(body)
		if err != nil {
			w.WriteHeader(500)
			logger.Error("failed to marshal taskCallbackRequest", err)
			return
		}

		callbackURL := strings.Replace(req.TaskDefinition.CompletionCallbackUrl, "https", "http", -1)
		res, err := http.Post(callbackURL, "application/json", bytes.NewBuffer(b))

		if err != nil || res.StatusCode < 200 || res.StatusCode >= 400 {
			w.WriteHeader(500)

			var statusCode int
			if res != nil {
				statusCode = res.StatusCode
			}

			logger.Error("received unexpected response from callback", err, lager.Data(map[string]interface{}{"statusCode": statusCode}))
			return
		}

		w.WriteHeader(200)
	}
}

type taskCallbackRequest struct {
	TaskGUID string `json:"task_guid"`
	Result   struct {
		TaskGUID          string            `json:"task_guid"`
		ExecutionMetadata string            `json:"execution_metadata"`
		ProcessTypes      map[string]string `json:"process_types"`
		LifecycleType     string            `json:"lifecycle_type"`
		LifecycleMetadata struct {
			DockerImage string `json:"docker_image"`
		} `json:"lifecycle_metadata"`
	} `json:"result"`
}
