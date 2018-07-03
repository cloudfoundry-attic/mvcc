package test_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/mvcc"
	"code.cloudfoundry.org/perm/pkg/perm"
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
	cc            *mvcc.MVCC
	fakeUaaServer *httptest.Server
	permServer    *mvcc.PermServer
	permClient    *perm.Client

	admin mvcc.User
	user  mvcc.User
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MVCC Test Suite")
}

var _ = BeforeSuite(func() {
	ccPath := os.Getenv("CLOUD_CONTROLLER_SRC_PATH")
	ccConfigPath := os.Getenv("CLOUD_CONTROLLER_CONFIG_PATH")

	var err error
	cc, err = mvcc.DialMVCC(
		mvcc.WithCloudControllerPath(ccPath),
		mvcc.WithCloudControllerConfigPath(ccConfigPath),
	)
	Expect(err).NotTo(HaveOccurred())

	fakeUaaServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/.well-known/openid-configuration":
			w.Write([]byte(fmt.Sprintf(`
{
  "issuer": "http://%s"
}`, req.Host)))
		default:
			out, err := httputil.DumpRequest(req, true)
			Expect(err).NotTo(HaveOccurred())
			Fail(fmt.Sprintf("unexpected request: %s", out))
		}
	}))

	err = fakeUaaServer.Listener.Close()
	Expect(err).NotTo(HaveOccurred())

	customListener, err := net.Listen("tcp", "localhost:6789")
	Expect(err).NotTo(HaveOccurred())

	fakeUaaServer.Listener = customListener
	fakeUaaServer.Start()

	permServerBinPath := os.Getenv("PERM_SERVER_BIN_PATH")
	permServerCertsPath := os.Getenv("PERM_SERVER_CERTS_PATH")

	permServer, err = mvcc.DialPermServer(
		mvcc.WithPermBinaryPath(permServerBinPath),
		mvcc.WithPermCertsPath(permServerCertsPath),
	)
	Expect(err).NotTo(HaveOccurred())

	permCACert, err := ioutil.ReadFile(filepath.Join(permServerCertsPath, "perm-server.crt"))
	Expect(err).NotTo(HaveOccurred())

	rootCAPool := x509.NewCertPool()
	ok := rootCAPool.AppendCertsFromPEM([]byte(permCACert))
	Expect(ok).To(BeTrue())

	permClient, err = perm.Dial(
		"localhost:3333",
		perm.WithTLSConfig(&tls.Config{
			RootCAs: rootCAPool,
		}),
	)
	Expect(err).NotTo(HaveOccurred())

	adminUUID, err := mvcc.RandomUUID("admin")
	Expect(err).NotTo(HaveOccurred())

	adminToken, err := createSignedToken(adminUUID, true)
	Expect(err).NotTo(HaveOccurred())

	admin = mvcc.User{
		UUID:        adminUUID,
		AccessToken: adminToken.AccessToken,
	}

	userUUID, err := mvcc.RandomUUID("user")
	Expect(err).NotTo(HaveOccurred())

	userToken, err := createSignedToken(userUUID, false)
	Expect(err).NotTo(HaveOccurred())

	user = mvcc.User{
		UUID:        userUUID,
		AccessToken: userToken.AccessToken,
	}
})

var _ = AfterSuite(func() {
	if cc != nil {
		err := cc.Kill()
		Expect(err).NotTo(HaveOccurred())
	}

	fakeUaaServer.Close()

	if permServer != nil {
		err := permServer.Kill()
		Expect(err).NotTo(HaveOccurred())
	}

	err := permClient.Close()
	Expect(err).NotTo(HaveOccurred())
})

var tokenRoles map[string]string = map[string]string{
	"admin": `{
		"jti": "9be1892c72a3472d8f80d11fc9825784",
		"sub": "%s",
		"scope": [
			"openid",
			"cloud_controller.admin",
			"cloud_controller.read",
			"cloud_controller.write"
		],
		"client_id": "cf",
		"cid": "cf",
		"azp": "cf",
		"grant_type": "password",
		"user_id": "%s",
		"origin": "uaa",
		"user_name": "admin",
		"email": "admin",
		"rev_sig": "666a6510",
		"zid": "uaa",
		"aud": [
			"cloud_controller",
			"password",
			"cf",
			"openid"
		],
		"iat": %d,
		"exp": %d,
		"iss": "%s"
	}`,
	"non-admin": `{
		"jti": "9be1892c72a3472d8f80d11fc9825784",
		"sub": "%s",
		"scope": [
			"openid",
			"cloud_controller.read",
			"cloud_controller.write"
		],
		"client_id": "cf",
		"cid": "cf",
		"azp": "cf",
		"grant_type": "password",
		"user_id": "%s",
		"origin": "uaa",
		"user_name": "non-admin",
		"email": "non-admin",
		"rev_sig": "666a6510",
		"zid": "uaa",
		"aud": [
			"cloud_controller",
			"password",
			"cf",
			"openid"
		],
		"iat": %d,
		"exp": %d,
		"iss": "%s"
	}`,
}

func createSignedToken(userId string, isAdmin bool) (*oauth2.Token, error) {
	issuedAt := time.Now().AddDate(-50, 0, 0).Unix() // 50 years ago
	expireAt := time.Now().AddDate(50, 0, 0)

	var template string
	if isAdmin {
		template = tokenRoles["admin"]
	} else {
		template = tokenRoles["non-admin"]
	}

	payload := fmt.Sprintf(template, userId, userId, issuedAt, expireAt.Unix(), validIssuer)

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

func randomName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().Nanosecond())
}
