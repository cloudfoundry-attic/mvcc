package diegox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"code.cloudfoundry.org/bbs/models"
)

type StagingServer struct {
	mux    *http.ServeMux
	server *http.Server
}

func NewStagingServer(...StagingServerOption) *StagingServer {

	mux := &http.ServeMux{}
	mux.HandleFunc("/v1/tasks/desire.r2", desireTaskHandler)

	return &StagingServer{
		mux: mux,
	}
}

func (s *StagingServer) ListenAndServe(addr string) error {
	if s.server == nil {
		s.server = &http.Server{
			Addr: addr,
		}
		s.server.Handler = s.mux
	}

	return s.server.ListenAndServe()
}

type StagingServerOption func(*stagingServerOptions)

type stagingServerOptions struct{}

func desireTaskHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	bits, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("FAILED TO READ?!?!?!?!")
		w.WriteHeader(413) // TODO
		return
	}


	req := &models.DesireTaskRequest{}
	if err = req.Unmarshal(bits); err != nil {
		fmt.Println("FAILED TO UNMARSHAL HERE")
		w.WriteHeader(414) // TODO
		return
	}

	body := &stagingCallbackRequest{}
	body.Result.LifecycleMetadata.DockerImage = "alpine"
	body.Result.LifecycleType = "docker"
	body.Result.ProcessTypes = map[string]string{
		"docker": "docker run",
	}
	body.Result.

	b, err := json.Marshal(body)
	if err != nil {
		fmt.Println("FAILED TO MARSHAL BODY")
		w.WriteHeader(415) // TODO
		return
	}

	callbackURL := strings.Replace(req.TaskDefinition.CompletionCallbackUrl, "https", "http", -1)
	res, err := http.Post(callbackURL, "application/json", bytes.NewBuffer(b))
	if err != nil || res.StatusCode < 200 || res.StatusCode >= 400 {
		fmt.Println("TRYING TO TALK TO", callbackURL)
		fmt.Println("FAIELD TO TALK TO CAPI HAPPILY", err, res)
		w.WriteHeader(416) // TODO
		return
	}
	fmt.Println("such great success")

	// proto := models.TaskLifecycleResponse{
	// Error: &models.Error{
	// Message: "this error is returned",
	// Type:    8,
	// },
	// }
	// b, err := proto.Marshal()
	// if err != nil {
	// fmt.Println("GREP THIS")
	// }
	w.WriteHeader(200)
	// fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

type stagingCallbackRequest struct {
	Result struct {
		ExecutionMetadata string            `json:"execution_metadata"`
		ProcessTypes      map[string]string `json:"process_types"`
		LifecycleType     string            `json:"lifecycle_type"`
		LifecycleMetadata struct {
			DockerImage string `json:"docker_image"`
		} `json:"lifecycle_metadata"`
	} `json:"result"`
}

// result: {
// execution_metadata: String,
// process_types:      dict(Symbol, String),
// lifecycle_type:     Lifecycles::DOCKER,
// lifecycle_metadata: {
// docker_image: String,
// }
// }
