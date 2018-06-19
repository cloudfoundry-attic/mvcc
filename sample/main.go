package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/mvcc"
)

func main() {
	cc, err := mvcc.Dial()
	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	defer cc.Kill()
}
