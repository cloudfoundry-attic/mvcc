package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"code.cloudfoundry.org/mvcc"
)

func main() {
	cc, err := mvcc.DialMVCC()
	if err != nil {
		fmt.Println("err:", err.Error())
		os.Exit(1)
	}

	res, err := http.Get("http://localhost:8181/v2/info")
	if err != nil {
		fmt.Println("err:", err.Error())
		os.Exit(1)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("err:", err.Error())
		os.Exit(1)
	}

	fmt.Printf("response from /v2/info:\n%s\n", body)

	defer cc.Kill()
}
