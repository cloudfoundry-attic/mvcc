package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

func main() {
	oauthServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/.well-known/openid-configuration":
			fmt.Println("req to openid-configuration")
			w.Write([]byte(fmt.Sprintf(`
{
  "issuer": "http://%s",
  "jwks_uri": "http://%s/token_keys"
}`, req.Host, req.Host)))
		case "/token_keys":
			fmt.Println("req to token_keys")
			w.Write([]byte(`
{
	"keys": [
		{
			"kty": "HSA",
			"e": "AQAB",
			"n": "APdfDb_pJpRhdbs9jDCWifExV81fROWxVVaLSh8R5dRxEIY8RRVS5AYmoI4etabv3gubYw3cgNIJGFqS3xLwJwGrVf7fh8YIIxm_H6QNG2mys7Rn80RufRpqrkas0EBcJqa_zQpS3QnINJ6ZkrSXhghYdD0R_01VQpQ7OnXFLKCAxUo7y0vUiMQhdKf0y8YhRd5v-cvujgze0vQnWrDQ9UY224OPnNtJK1zv2E7Ssn43PTEt1OxF2lYLuSqJUw8lEiE8FTQIBIUj0yqfiMQ8dn4GeJem8nTfsRyNHBOHF-HddJW-RrQ-ryvFLLpFu0H0wecSzlF-5SXOsTUpGuMQ0pU"
		}
	]
}`))
		default:
			fmt.Printf("unhandled request: %#v", req)
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Printf("body: %s", data)
			w.Write([]byte("{}"))
		}
	}))

	err := oauthServer.Listener.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	customListener, err := net.Listen("tcp", "localhost:6789")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	oauthServer.Listener = customListener
	oauthServer.Start()

	fmt.Println("listening on :6789")

	time.Sleep(30 * time.Minute)
}
