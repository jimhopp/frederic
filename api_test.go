package frederic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"appengine"
	"appengine/aetest"
	"appengine/datastore"
	"appengine/user"
)

func TestAddClient(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	data := strings.NewReader(`{"Firstname": "frederic", "Lastname": "ozanam"}`)
	req, err := inst.NewRequest("PUT", "/api/addclient", data)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)

	addclient(c, w, req)

	code := w.Code
	if code != http.StatusCreated {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body := w.Body.Bytes()
	if !bytes.Contains(body, []byte(`{"Firstname":"frederic","Lastname":"ozanam"}`)) {
		t.Errorf("got body %v (%v), want %v", body, string(body), []byte(`{"Firstname":"frederic","Lastname":"ozanam"}`))
	}

	q := datastore.NewQuery("SVDPClient")
	clients := make([]client, 0, 10)
	if _, err := q.GetAll(c, &clients); err != nil {
		t.Fatalf("error on GetAll: %v", err)
		return
	}
	if len(clients) != 1 {
		t.Errorf("got %v records in query, expected %v", len(clients), 1)
	}
}

func TestGetAllClients(t *testing.T) {
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
	id := make([]int64, 3)
	for i := 0; i < len(newclients); i++ {
		id[i], err = addclienttodb(newclients[i], inst)
		log.Printf("TestAllClients: got %v from addclienttodb\n", id)
		if err != nil {
			t.Fatalf("unable to add client: %v", err)
		}
	}

	req, err := inst.NewRequest("GET", "/api/getallclients", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	aetest.Login(&user.User{Email: "test@example.org"}, req)
	w := httptest.NewRecorder()

	c := appengine.NewContext(req)
	getallclients(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	createdclientrecs := []clientrec{}
	err = json.Unmarshal(body, &createdclientrecs)
	if err != nil {
		t.Errorf("error unmarshaling response to getclients %v\n", err)
	}
	if len(createdclientrecs) != len(newclients) {
		t.Errorf("got %v clients, want %v\n", len(createdclientrecs),
			len(newclients))
	}
	for i := 0; i < len(newclients); i++ {
		found := false
		for j := 0; j < len(createdclientrecs); j++ {
			if reflect.DeepEqual(createdclientrecs[j].Clt,
				newclients[i]) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unable to find %v in %v",
				newclients[i], &createdclientrecs)
		}
	}
}

func addclienttodb(clt client, inst aetest.Instance) (id int64, err error) {
	data, err := json.Marshal(clt)
	if err != nil {
		return -1, err
	}

	req, err := inst.NewRequest("PUT", "/api/addclient", bytes.NewReader(data))
	if err != nil {
		return -1, err
	}
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)

	addclient(c, w, req)

	code := w.Code
	if code != http.StatusCreated {
		return -1, errors.New(fmt.Sprintf("got code on addclient %v, want %v",
			code, http.StatusCreated))
	}

	body := make([]byte, w.Body.Len())
	_, err = w.Body.Read(body)
	newrec := &clientrec{}
	err = json.Unmarshal(body, newrec)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("unable to parse %v, got err %v",
			string(body), err))
	}
	return newrec.Id, nil
}
