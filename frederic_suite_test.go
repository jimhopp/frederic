package frederic_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"

	"testing"
)

func TestFrederic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Frederic Suite")
}

var agoutiDriver *agouti.WebDriver

var _ = BeforeSuite(func() {
	// Choose a WebDriver:

	//agoutiDriver = agouti.PhantomJS()
	//agoutiDriver = agouti.Selenium()
	 agoutiDriver = agouti.ChromeDriver()

	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(agoutiDriver.Stop()).To(Succeed())
})
