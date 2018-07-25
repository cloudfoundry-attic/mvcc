package diegox

import (
	"fmt"
	"log"
	"net/http"
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
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	server := NewStagingServer()
	log.Fatal(server.ListenAndServe(":8889"))
}
