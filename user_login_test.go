package frederic_test

import (
//	"appengine/aetest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
//	. "github.com/sclevine/agouti/matchers"
)

var _ = Describe("UserLogin", func() {
	var page *agouti.Page

	BeforeEach(func() {
//inst, _ := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
//        defer inst.Close()
		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})
	It("should manage user authentication", func() {
		By("redirecting the user to the login form from the home page", func() {
//			Expect(page.Navigate("http://localhost:8080")).To(Succeed())
//			Expect(page).To(HaveURL("http://localhost:8080/login"))
		})

	})
})
