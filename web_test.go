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
	{"/api/visit/", false, visitrouter, http.StatusUnauthorized},
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
	{"/editvisit/", true, editvisitpage, http.StatusFound},
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
	}
	for i := 0; i < len(rows); i++ {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("got body %v, did not contain %v", string(body), rows[i])
		}
	}
}

func TestEditVisitPageForNonexistentClient(t *testing.T) {
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

	url := "/editvisit/" + sid + "/12345/edit"
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", true)

	editvisitpage(c, w, req)

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

func TestEditVisitPageMissingPathInfo(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	url := "/editvisit/"
	req, err := inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)

	addTestUser(c, "test@example.org", true)

	editvisitpage(c, w, req)

	code := w.Code
	if code != http.StatusBadRequest {
		t.Errorf("got code %v, want %v", code, http.StatusBadRequest)
	}

	body := w.Body.Bytes()
	expected := []byte("id is missing in path for update request /editvisit/")
	if !bytes.Contains(body, expected) {
		t.Errorf("got body %v, did not contain %v", string(body),
			string(expected))
	}

	url += "12345"
	req, err = inst.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w = httptest.NewRecorder()
	c = appengine.NewContext(req)

	editvisitpage(c, w, req)

	code = w.Code
	if code != http.StatusBadRequest {
		t.Errorf("got code %v, want %v", code, http.StatusBadRequest)
	}

	body = w.Body.Bytes()
	expected = []byte("id is missing in path for update request /editvisit/")
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

func TestEditVisitPage(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}
	cltid, err := addclienttodb(newclient, inst)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	visits := []visit{
		{Vincentians: "Michael, Mary Margaret",
			Visitdate:           "2013-02-03",
			Assistancerequested: "test1"},
		{Vincentians: "Irene, Jim",
			Visitdate:           "2013-01-03",
			Assistancerequested: "test2"},
	}
	visitIds := make([]int64, len(visits))
	for i, vst := range visits {
		data, err := json.Marshal(vst)
		if err != nil {
			t.Fatalf("Failed to marshal %v", visits[i])
		}
		req, err := inst.NewRequest("PUT", "/api/visit/"+
			strconv.FormatInt(cltid, 10), bytes.NewReader(data))
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

		visitrouter(c, w, req)

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
		visitIds[i] = newrec.Id
	}

	req, err := inst.NewRequest("GET", "/editvisit/"+strconv.FormatInt(cltid, 10)+
		"/"+strconv.FormatInt(visitIds[0], 10)+"/edit", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	aetest.Login(&user.User{Email: "test@example.org"}, req)
	w := httptest.NewRecorder()

	c := appengine.NewContext(req)
	editvisitpage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	if !bytes.Contains(body, []byte(visits[0].Assistancerequested)) {
		t.Errorf("unable to find %v in %v",
			visits[0].Assistancerequested, string(body))
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
				Visitdate:           "2013-02-03",
				Assistancerequested: "test1"},
			{Vincentians: "Irene, Jim",
				Visitdate:           "2013-01-03",
				Assistancerequested: "test2"},
		},
		{
			{Vincentians: "Eileen, Lynn",
				Visitdate:           "2013-04-03",
				Assistancerequested: "test3"},
			{Vincentians: "Stu & Anne",
				Visitdate:           "2013-03-03",
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

	req, err := inst.NewRequest("GET", "/visits?startdate=2013-01-03&enddate=2013-04-03", nil)
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
	m, err := regexp.Match(`(?s).*2013-04-03.*2013-03-03.*2013-02-03.*2013-01-03.*`, body)
	if err != nil {
		t.Errorf("got error on regexp match: %v", err)
	}
	if !m {
		t.Errorf("visit dates not sorted: %v", string(body))
	}
}

func TestDownloadVisitsInRange(t *testing.T) {
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
				Visitdate:           "2013-02-03",
				Assistancerequested: "test1"},
			{Vincentians: "Irene, Jim",
				Visitdate:           "2013-01-03",
				Assistancerequested: "test2"},
		},
		{
			{Vincentians: "Eileen, Lynn",
				Visitdate:           "2013-04-03",
				Assistancerequested: "test3"},
			{Vincentians: "Stu & Anne",
				Visitdate:           "2013-03-03",
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

	req, err := inst.NewRequest("GET", "/visits?startdate=2013-01-03&enddate=2013-04-03&csv=true", nil)
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

	headers := w.HeaderMap
	c.Infof("headers=%v", headers)
	if headers["Content-Type"][0] != "text/csv" {
		t.Errorf("expected Content-Type to contain text/csv but it is %v", headers)
	}
	body := w.Body.Bytes()
	for i, viz := range visits {
		for j, vst := range viz {
			if !bytes.Contains(body, []byte(vst.Assistancerequested)) {
				t.Errorf("unable to find %v in %v",
					vst, string(body))
			}
			clientName := []byte(`"` + newclients[i].Firstname + ` ` + newclients[i].Lastname + `"`)
			if !bytes.Contains(body, clientName) {
				t.Errorf("CSV unable to find %v in %v", string(clientName), string(body))
			}
			visitData := []byte(`"` + vst.Visitdate + `","` +
				vst.Vincentians + `","` +
				vst.Assistancerequested + `","` +
				vst.Giftcardamt + `","` +
				vst.Numfoodboxes + `","` +
				vst.Rentassistance + `","` +
				vst.Utilitiesassistance + `","` +
				vst.Waterbillassistance + `","` +
				vst.Otherassistancetype + `","` +
				vst.Otherassistanceamt + `","` +
				vst.Vouchersclothing + `","` +
				vst.Vouchersfurniture + `","` +
				vst.Vouchersother + `","` +
				vst.Comment + `"`)
			if !bytes.Contains(body, visitData) {
				t.Errorf("visit[%v,%v]: CSV unable to find %v in %v", i, j, string(visitData), string(body))
			}
		}
	}
	m, err := regexp.Match(`(?s).*2013-04-03.*2013-03-03.*2013-02-03.*2013-01-03.*`, body)
	if err != nil {
		t.Errorf("got error on regexp match: %v", err)
	}
	if !m {
		t.Errorf("visit dates not sorted: %v", string(body))
	}
}
