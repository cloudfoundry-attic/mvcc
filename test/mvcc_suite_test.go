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
	"strconv"
	"time"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/mvcc"
	"code.cloudfoundry.org/mvcc/fixtures"
	"code.cloudfoundry.org/perm/pkg/api"
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
	permListener  net.Listener
	permServer    *api.Server
	permClient    *perm.Client

	admin mvcc.User
	user  mvcc.User

	actor perm.Actor
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MVCC Test Suite")
}

var _ = BeforeSuite(func() {
	permServerCert, err := tls.X509KeyPair([]byte(fixtures.TLSCertificate), []byte(fixtures.TLSKey))
	Expect(err).NotTo(HaveOccurred())

	permServer = api.NewServer(api.WithTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{permServerCert},
	}), api.WithLogger(lagertest.NewTestLogger("perm")))

	permListener, err = net.Listen("tcp", "localhost:0")
	Expect(err).NotTo(HaveOccurred())

	go func() {
		err = permServer.Serve(permListener)
		Expect(err).NotTo(HaveOccurred())
	}()

	permCA := []byte(fixtures.TLSCertificateAuthority)

	rootCAPool := x509.NewCertPool()
	ok := rootCAPool.AppendCertsFromPEM(permCA)
	Expect(ok).To(BeTrue())

	permClient, err = perm.Dial(
		permListener.Addr().String(),
		perm.WithTLSConfig(&tls.Config{
			RootCAs: rootCAPool,
		}),
	)
	Expect(err).NotTo(HaveOccurred())

	_, p, err := net.SplitHostPort(permListener.Addr().String())
	Expect(err).NotTo(HaveOccurred())

	permPort, err := strconv.ParseInt(p, 0, 0)
	Expect(err).NotTo(HaveOccurred())

	permCAFile, err := ioutil.TempFile("", "ca")
	Expect(err).NotTo(HaveOccurred())

	defer os.Remove(permCAFile.Name())

	_, err = permCAFile.Write(permCA)
	Expect(err).NotTo(HaveOccurred())

	cc, err = mvcc.DialMVCC(
		mvcc.WithPerm(mvcc.PermOptions{
			Port:       int(permPort),
			CACertPath: permCAFile.Name(),
		}),
	)
	Expect(err).NotTo(HaveOccurred())

	adminUUID := mvcc.RandomUUID("admin")

	adminToken, err := createSignedToken(adminUUID, true)
	Expect(err).NotTo(HaveOccurred())

	admin = mvcc.User{
		UUID:        adminUUID,
		AccessToken: adminToken.AccessToken,
	}

	userUUID := mvcc.RandomUUID("user")

	userToken, err := createSignedToken(userUUID, false)
	Expect(err).NotTo(HaveOccurred())

	user = mvcc.User{
		UUID:        userUUID,
		AccessToken: userToken.AccessToken,
	}

	actor = perm.Actor{
		ID:        user.UUID,
		Namespace: validIssuer,
	}
})

var _ = AfterSuite(func() {
	if cc != nil {
		err := cc.Kill()
		Expect(err).NotTo(HaveOccurred())
	}

	permServer.GracefulStop()

	err := permClient.Close()
	Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
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

	err := fakeUaaServer.Listener.Close()
	Expect(err).NotTo(HaveOccurred())

	customListener, err := net.Listen("tcp", "localhost:6789")
	Expect(err).NotTo(HaveOccurred())

	fakeUaaServer.Listener = customListener
	fakeUaaServer.Start()
})

var _ = AfterEach(func() {
	fakeUaaServer.Close()
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
