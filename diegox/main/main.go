package main

import (
	"log"

	"code.cloudfoundry.org/mvcc/diegox"
)

func main() {
	server := diegox.NewBBSServer()
	log.Fatal(server.ListenAndServe(":8889"))
}
