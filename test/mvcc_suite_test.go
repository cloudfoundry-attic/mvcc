package test_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"time"

	"code.cloudfoundry.org/mvcc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"
	jose "gopkg.in/square/go-jose.v2"

	"testing"
)

const (
	signingKey  = "tokensecret"
	validIssuer = "http://localhost:6789"
)

var (
	cc          *mvcc.MVCC
	oauthServer *httptest.Server
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MVCC Test Suite")
}

var _ = BeforeSuite(func() {
	ccPath := os.Getenv("CLOUD_CONTROLLER_SRC_PATH")
	ccConfigPath := os.Getenv("CLOUD_CONTROLLER_CONFIG_PATH")

	var err error
	cc, err = mvcc.Dial(
		mvcc.WithCloudControllerPath(ccPath),
		mvcc.WithCloudControllerConfigPath(ccConfigPath),
	)
	Expect(err).NotTo(HaveOccurred())

	oauthServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/.well-known/openid-configuration":
			w.Write([]byte(fmt.Sprintf(`
{
  "issuer": "http://%s",
  "jwks_uri": "http://%s/token_keys"
}`, req.Host, req.Host)))
		default:
			out, err := httputil.DumpRequest(req, true)
			Expect(err).NotTo(HaveOccurred())
			Fail(fmt.Sprintf("unexpected request: %s", out))
		}
	}))

	err = oauthServer.Listener.Close()
	Expect(err).NotTo(HaveOccurred())

	customListener, err := net.Listen("tcp", "localhost:6789")
	Expect(err).NotTo(HaveOccurred())

	oauthServer.Listener = customListener
	oauthServer.Start()
})

var _ = AfterSuite(func() {
	if cc != nil {
		err := cc.Kill()
		Expect(err).NotTo(HaveOccurred())
	}

	oauthServer.Close()
})

func createSignedToken() (*oauth2.Token, error) {
	issuedAt := time.Now().AddDate(-50, 0, 0).Unix() // 50 years ago
	expireAt := time.Now().AddDate(50, 0, 0)
	payload := fmt.Sprintf(`
{
  "jti": "9be1892c72a3472d8f80d11fc9825784",
  "sub": "4d3e04b1-f89f-4370-ada7-70e8d1b7f3c1",
  "scope": [
    "cloud_controller.admin",
    "password.write",
    "openid",
    "uaa.user"
  ],
  "client_id": "cf",
  "cid": "cf",
  "azp": "cf",
  "grant_type": "password",
  "user_id": "4d3e04b1-f89f-4370-ada7-70e8d1b7f3c1",
  "origin": "uaa",
  "user_name": "admin",
  "email": "admin",
  "rev_sig": "666a6510",
  "zid": "uaa",
  "aud": [
    "cloud_controller",
    "password",
    "cf",
    "uaa",
    "openid"
  ],
	"iat": %d,
	"exp": %d,
	"iss": "%s"
}`, issuedAt, expireAt.Unix(), validIssuer)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte(signingKey)}, nil)
	Expect(err).NotTo(HaveOccurred())

	signedToken, err := signer.Sign([]byte(payload))
	Expect(err).NotTo(HaveOccurred())

	serialized := signedToken.FullSerialize()

	var token struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	}

	err = json.Unmarshal([]byte(serialized), &token)
	Expect(err).NotTo(HaveOccurred())

	return &oauth2.Token{
		AccessToken: fmt.Sprintf("bearer %s.%s.%s", token.Protected, token.Payload, token.Signature),
		Expiry:      expireAt,
	}, nil
}
