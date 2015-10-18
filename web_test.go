package frederic

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"appengine"
	"appengine/aetest"
	"appengine/user"
)

type EndpointTest struct {
	url      string
	handler  func(appengine.Context, http.ResponseWriter, *http.Request)
	expected int
}

var endpoints = []EndpointTest{
	{"/api/client", addclient, http.StatusUnauthorized},
	{"/api/client/", editclient, http.StatusUnauthorized},
	{"/api/visit/", addvisit, http.StatusUnauthorized},
	{"/api/getallclients", getallclients, http.StatusUnauthorized},
	{"/api/getallvisits/", getallvisits, http.StatusUnauthorized},
	{"/", homepage, http.StatusFound},
	{"/listclients", listclientspage, http.StatusFound},
	{"/client", getclientpage, http.StatusFound},
	{"/newclient", newclientpage, http.StatusFound},
	{"/editclient", editclientpage, http.StatusFound},
	{"/recordvisit/", recordvisitpage, http.StatusFound},
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
	addTestUser(c, "test@example.org")

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
	addTestUser(c, "test@example.org")

	listclientspage(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	rows := []string{"<td>Clients</td>",
		"<a href=\"/client/" + strconv.FormatInt(ids[0], 10) +
			"\">frederic ozanam</a>",
		"<a href=\"/client/" + strconv.FormatInt(ids[1], 10) +
			"\">John Doe</a>",
		"<a href=\"/client/" + strconv.FormatInt(ids[2], 10) +
			"\">Jane Doe</a>",
	}
	for i := 0; i < len(rows); i++ {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("got body %v, did not contain %v", string(body), rows[i])
		}
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
	addTestUser(c, "test@example.org")

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
	addTestUser(c, "test@example.org")

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
	addTestUser(c, "test@example.org")

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
	addTestUser(c, "test@example.org")

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
	addTestUser(c, "test@example.org")

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
	addTestUser(c, "test@example.org")

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
	addTestUser(c, "test@example.org")

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

	addTestUser(c, "test@example.org")

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
