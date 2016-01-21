package frederic

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"appengine"
	"appengine/aetest"
	"appengine/user"
)

type EndpointTest struct {
	url           string
	humanReadable bool
	handler       func(appengine.Context, http.ResponseWriter, *http.Request)
	expected      int
}

var endpoints = []EndpointTest{
	{"/api/client", false, addclient, http.StatusUnauthorized},
	{"/api/client/", false, editclient, http.StatusUnauthorized},
	{"/api/visit/", false, addvisit, http.StatusUnauthorized},
	{"/api/getallclients", false, getallclients, http.StatusUnauthorized},
	{"/api/getallvisits/", false, getallvisits, http.StatusUnauthorized},
	{"/api/getvisitsinrange/", false, getvisitsinrange, http.StatusUnauthorized},
	{"/api/users/", false, getallusers, http.StatusUnauthorized},
	{"/api/users/edit", false, editusers, http.StatusUnauthorized},
	{"/", true, homepage, http.StatusFound},
	{"/visits", true, listvisitsinrangepage, http.StatusFound},
	{"/clients", true, listclientspage, http.StatusFound},
	{"/client", true, getclientpage, http.StatusFound},
	{"/newclient", true, newclientpage, http.StatusFound},
	{"/editclient", true, editclientpage, http.StatusFound},
	{"/recordvisit/", true, recordvisitpage, http.StatusFound},
}

func TestHomePage(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create req1: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	homepage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	expected := []byte("test@example.org")
	if !bytes.Contains(body, expected) {
		t.Errorf("got body %v, did not contain %v", string(body), string(expected))
	}
	if !bytes.Contains(body, []byte("Logout")) {
		t.Errorf("got body %v, did not contain %v", body,
			[]byte("Logout"))
	}
}

func TestListClientsPage(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclients := []client{
		{Firstname: "frederic", Lastname: "ozanam"},
		{Firstname: "John", Lastname: "Doe"},
		{Firstname: "Jane", Lastname: "Doe"},
	}
	ids := make([]int64, len(newclients))
	for i := 0; i < len(newclients); i++ {
		id, err := addclienttodb(newclients[i], inst)
		if err != nil {
			t.Fatalf("unable to add client: %v", err)
		}
		ids[i] = id
	}
	req, err := inst.NewRequest("GET", "/listclients", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	listclientspage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	rows := []string{"<td>Clients</td>",
		"<a href=\"/client/" + strconv.FormatInt(ids[0], 10) +
			"\">ozanam, frederic</a>",
		"<a href=\"/client/" + strconv.FormatInt(ids[1], 10) +
			"\">Doe, John</a>",
		"<a href=\"/client/" + strconv.FormatInt(ids[2], 10) +
			"\">Doe, Jane</a>",
	}
	for i := 0; i < len(rows); i++ {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("got body %v, did not contain %v", string(body), rows[i])
		}
	}
}

func TestListClientsPageIsSorted(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclients := []client{
		{Firstname: "frederic", Lastname: "ozanam"},
		{Firstname: "John", Lastname: "Doe"},
		{Firstname: "Jane", Lastname: "Doe"},
	}
	ids := make([]int64, 3)
	for i := 0; i < len(newclients); i++ {
		id, err := addclienttodb(newclients[i], inst)
		if err != nil {
			t.Fatalf("unable to add client: %v", err)
		}
		ids[i] = id
	}
	req, err := inst.NewRequest("GET", "/listclients", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	listclientspage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	m, err := regexp.Match(`(?s).*Doe.*Doe.*ozanam.*`, body)

	if err != nil {
		t.Errorf("got error on regexp match: %v", err)
	}
	if !m {
		t.Errorf("names not sorted: %v", string(body))
	}
}

func TestGetClientPage(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}

	id, err := addclienttodb(newclient, inst)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	sid := strconv.FormatInt(id, 10)

	url := "/client/" + sid
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	getclientpage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	rows := []string{`value="frederic"`,
		`value="ozanam"`,
		"<a href=\"/editclient/" + sid + "\">(edit)</a>",
	}
	for i := 0; i < len(rows); i++ {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("got body %v, did not contain %v", string(body), rows[i])
		}
	}
}

func TestGetClientNotFound(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	url := "/client/1234"
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	getclientpage(c, w, req)

	code := w.Code
	if code != http.StatusNotFound {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	msg := []byte("unable to find client")
	if !bytes.Contains(body, msg) {
		t.Errorf("got body %v, did not contain %v", string(body),
			string(msg))
	}
}

func TestGetClientMissingParm(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	url := "/client/"
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	getclientpage(c, w, req)

	code := w.Code
	if code != http.StatusNotFound {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	msg := []byte("client id missing in path")
	if !bytes.Contains(body, msg) {
		t.Errorf("got body %v, did not contain %v", string(body),
			string(msg))
	}
}

func TestEditClientPage(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}

	id, err := addclienttodb(newclient, inst)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	sid := strconv.FormatInt(id, 10)

	url := "/editclient/" + sid
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	editclientpage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	rows := []string{`value="frederic"`,
		`value="ozanam"`,
		`method: "PUT"`,
		`url: "/api/client/` + sid + `"`,
	}
	for i := 0; i < len(rows); i++ {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("got body %v, did not contain %v", string(body), rows[i])
		}
	}
}

func TestAddClientPage(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/addclient", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	newclientpage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	rows := []string{`method: "POST"`,
		`url: "/api/client"`,
	}
	for i := 0; i < len(rows); i++ {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("got body %v, did not contain %v", string(body), rows[i])
		}
	}
	//TODO: confirm response, create new req with filled-in values, submit?
	//      Or does this call for something like Selenium?
}

func TestEndpointsNotAuthenticated(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	for i := 0; i < len(endpoints); i++ {
		req, err := inst.NewRequest("GET", endpoints[i].url, nil)
		if err != nil {
			t.Fatalf("Failed to create req1: %v", err)
		}
		w := httptest.NewRecorder()
		c := appengine.NewContext(req)

		endpoints[i].handler(c, w, req)

		code := w.Code
		if code != endpoints[i].expected {
			t.Errorf("got code %v for endpoint %v, want %v", code,
				endpoints[i].url, endpoints[i].expected)
		}
	}
}

func TestEndpointsNotAuthorized(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	for i := 0; i < len(endpoints); i++ {
		req, err := inst.NewRequest("GET", endpoints[i].url, nil)
		if err != nil {
			t.Fatalf("Failed to create req1: %v", err)
		}
		w := httptest.NewRecorder()
		c := appengine.NewContext(req)

		aetest.Login(&user.User{Email: "test@example.org"}, req)
		endpoints[i].handler(c, w, req)

		code := w.Code
		if code != http.StatusForbidden {
			t.Errorf("got code %v for endpoint %v, want %v", code,
				endpoints[i].url, http.StatusForbidden)
		}
		if endpoints[i].humanReadable {
			body := w.Body.Bytes()
			notauth := `Sorry`
			if !bytes.Contains(body, []byte(notauth)) {
				t.Errorf("endpoint %v: got body %v, did not contain %v", endpoints[i].url, string(body), notauth)
			}
		}
	}
}

func TestAddVisitPage(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}

	id, err := addclienttodb(newclient, inst)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	sid := strconv.FormatInt(id, 10)

	url := "/recordvisit/" + sid
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	recordvisitpage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	rows := []string{`frederic`,
		`ozanam`,
		`method: "PUT"`,
		`url: "/api/visit/`,
	}
	for i := 0; i < len(rows); i++ {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("got body %v, did not contain %v", string(body), rows[i])
		}
	}
}

func TestAddVisitPageForNonexistentClient(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}

	id, err := addclienttodb(newclient, inst)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	id++

	sid := strconv.FormatInt(id, 10)

	url := "/recordvisit/" + sid
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	recordvisitpage(c, w, req)

	code := w.Code
	if code != http.StatusNotFound {
		t.Errorf("got code %v, want %v", code, http.StatusNotFound)
	}

	body := w.Body.Bytes()
	expected := []byte("Unable to find client with id " + sid)
	if !bytes.Contains(body, expected) {
		t.Errorf("got body %v, did not contain %v", string(body),
			string(expected))
	}
}

func TestAddVisitPageMissingClient(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	url := "/recordvisit/"
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)

	addTestUser(c, "test@example.org", true)

	recordvisitpage(c, w, req)

	code := w.Code
	if code != http.StatusBadRequest {
		t.Errorf("got code %v, want %v", code, http.StatusNotFound)
	}

	body := w.Body.Bytes()
	expected := []byte("Invalid/missing client id")
	if !bytes.Contains(body, expected) {
		t.Errorf("got body %v, did not contain %v", string(body),
			string(expected))
	}
}

func TestListUsersPage(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newusers := []appuser{
		{Email: "frederic@example.org", IsAdmin: true},
		{Email: "j@example.org", IsAdmin: false},
		{Email: "x@example.org", IsAdmin: false},
	}

	req, err := inst.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	for i := range newusers {
		_, err := addTestUser(c, newusers[i].Email,
			newusers[i].IsAdmin)
		if err != nil {
			t.Fatalf("unable to add user: %v", err)
		}
	}

	edituserspage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	rows := []string{"frederic@example.org",
		"j@example.org",
		"x@example.org",
		`<input type="checkbox" id="admin0" name="admin" checked="checked">`,
		`<input type="checkbox" id="admin1" name="admin">`,
		// test@example.org is user #3
		`<input type="checkbox" id="admin3" name="admin">`,
	}
	for i := range rows {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("got body %v, did not contain %v", string(body), rows[i])
		}
	}
}

func TestListUsersPageNotAdmin(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", false)

	edituserspage(c, w, req)

	code := w.Code
	if code != http.StatusForbidden {
		t.Errorf("got code %v, want %v", code, http.StatusForbidden)
	}

}

func TestListVisitsInRange(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclients := []client{{Firstname: "frederic", Lastname: "ozanam"},
		{Firstname: "Elizabeth", Lastname: "Seton"},
	}
	cltids := make([]int64, len(newclients))
	for i, newclient := range newclients {
		cltids[i], err = addclienttodb(newclient, inst)
		log.Printf("TestAllVisits: got %v from addclienttodb\n", cltids[i])
		if err != nil {
			t.Fatalf("unable to add client: %v", err)
		}
	}

	visits := [][]visit{
		{
			{Vincentians: "Michael, Mary Margaret",
				Visitdate:           "2013-04-03",
				Assistancerequested: "test1"},
			{Vincentians: "Irene, Jim",
				Visitdate:           "2013-05-03",
				Assistancerequested: "test2"},
		},
		{
			{Vincentians: "Eileen, Lynn",
				Visitdate:           "2013-06-03",
				Assistancerequested: "test3"},
			{Vincentians: "Stu & Anne",
				Visitdate:           "2013-07-03",
				Assistancerequested: "test4"},
		},
	}

	numvisits := 0
	for i, viz := range visits {
		for _, vst := range viz {
			data, err := json.Marshal(vst)
			if err != nil {
				t.Fatalf("Failed to marshal %v", visits[i])
			}
			req, err := inst.NewRequest("PUT", "/visit/"+
				strconv.FormatInt(cltids[i], 10), bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Failed to create req: %v", err)
			}
			req.Header = map[string][]string{
				"Content-Type": {"application/json"},
			}
			aetest.Login(&user.User{Email: "test@example.org"}, req)

			w := httptest.NewRecorder()
			c := appengine.NewContext(req)
			addTestUser(c, "test@example.org", true)

			addvisit(c, w, req)

			code := w.Code
			if code != http.StatusOK {
				t.Errorf("got code %v, want %v", code, http.StatusCreated)
			}

			body := w.Body.Bytes()
			newrec := &visitrec{}
			err = json.Unmarshal(body, newrec)
			if err != nil {
				t.Errorf("unable to parse %v: %v", string(body), err)
			}
			numvisits++
		}
	}

	req, err := inst.NewRequest("GET", "/listvisitsinrange/", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	aetest.Login(&user.User{Email: "test@example.org"}, req)
	w := httptest.NewRecorder()

	c := appengine.NewContext(req)
	listvisitsinrangepage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	for _, viz := range visits {
		for _, vst := range viz {
			if !bytes.Contains(body, []byte(vst.Assistancerequested)) {
				t.Errorf("unable to find %v in %v",
					vst, string(body))
			}
		}
	}
}
