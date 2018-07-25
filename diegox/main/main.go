package main

import (
	"log"

	"code.cloudfoundry.org/mvcc/diegox"
)

func main() {
	server := diegox.NewStagingServer()
	log.Fatal(server.ListenAndServe(":8889"))
}
