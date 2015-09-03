package frederic_test

import (
/* OMG this actually works! But it does require
   * selenium server be running at the URL below in the NewPage call ('selenium-server')
   * local go_appengine server is running at the URL in the Navigate call 
   * Chromedriver is installed and on the path
   
   You can un-comment the aetest stuff, and that will start a test server. BUT
   I haven't figured out how to tell it which port to listen on, or alternatively 
   to determine which port it is actually running on.

*/
 
//	"appengine/aetest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

var _ = Describe("UserLogin", func() {
	var page *agouti.Page

	BeforeEach(func() {
//inst, _ := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
//        defer inst.Close()
		var err error
		page, err = agouti.NewPage("http://127.0.0.1:4444/wd/hub", agouti.Browser("chrome"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})
	It("should manage user authentication", func() {
		By("redirecting the user to the login form from the home page", func() {
			Expect(page.Navigate("http://localhost:8080")).To(Succeed())
			Expect(page).To(HaveURL("http://localhost:8080/_ah/login?continue=http%3A//localhost%3A8080/"))
		})

	})
})
