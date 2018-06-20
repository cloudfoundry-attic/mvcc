package test_test

import (
	"os"

	"code.cloudfoundry.org/mvcc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	cc *mvcc.MVCC
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MVCC Test Suite")
}

var _ = BeforeSuite(func() {
	ccPath := os.Getenv("CLOUD_CONTROLLER_SRC_PATH")
	Expect(ccPath).NotTo(Equal(""), "Must set CLOUD_CONTROLLER_SRC_PATH")

	ccConfigPath := os.Getenv("CLOUD_CONTROLLER_CONFIG_PATH")
	Expect(ccConfigPath).NotTo(Equal(""), "Must set CLOUD_CONTROLLER_CONFIG_PATH")

	var err error
	cc, err = mvcc.Dial(
		mvcc.WithCloudControllerPath(ccPath),
		mvcc.WithCloudControllerConfigPath(ccConfigPath),
	)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if cc != nil {
		err := cc.Kill()
		Expect(err).NotTo(HaveOccurred())
	}
})
