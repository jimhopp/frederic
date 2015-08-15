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
	"strconv"
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

	data := strings.NewReader(`{"Firstname": "frederic", "Lastname": "ozanam","Address":"123 Easy St","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212"}`)
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
	expected := []byte(`{"Firstname":"frederic","Lastname":"ozanam","Address":"123 Easy St","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Fammbrs":null}`)
	if !bytes.Contains(body, expected) {
		t.Errorf("got body %v (%v), want %v", body, string(body),
			expected)
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

func TestUpdateClient(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}
	id, err := addclienttodb(newclient, inst)
	log.Printf("TestUpdateClient: got %v from addclienttodb\n", id)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	data := strings.NewReader(`{"Id": ` + strconv.FormatInt(id, 10) +
		`, "Clt": {"Firstname": "Frederic", "Lastname": "Ozanam"}}`)
	req, err := inst.NewRequest("POST", "/api/editclient", data)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)

	editclient(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body := w.Body.Bytes()
	expected := []byte(`{"Firstname":"Frederic","Lastname":"Ozanam","Address":"","Apt":"","DOB":"","Phonenum":"","Fammbrs":null}`)
	if !bytes.Contains(body, expected) {
		t.Errorf("got body %v (%v), want %v", body, string(body),
			expected)
	}

	key := datastore.NewKey(c, "SVDPClient", "", id, nil)
	clt := client{}
	if err := datastore.Get(c, key, &clt); err != nil {
		t.Fatalf("error on Get: %v", err)
		return
	}
	expectedc := &client{}
	expectedc.Firstname = "Frederic"
	expectedc.Lastname = "Ozanam"
	if reflect.DeepEqual(clt, expectedc) {
		t.Errorf("db record shows %v, want %v", clt, expectedc)
	}
}

func TestUpdateErrorClient(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}
	id, err := addclienttodb(newclient, inst)
	log.Printf("TestUpdateClient: got %v from addclienttodb\n", id)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	data := strings.NewReader(`{"Id": ` + strconv.FormatInt(id, 10) +
		`, "Clt": {"Firstname": "Frederic", "Lastname": "Ozanam", "DOB":"alphabet"}}`)
	req, err := inst.NewRequest("POST", "/api/editclient", data)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)

	editclient(c, w, req)

	code := w.Code
	if code != http.StatusInternalServerError {
		t.Errorf("got code %v, want %v", code, http.StatusInternalServerError)
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
