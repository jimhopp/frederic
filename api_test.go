package frederic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
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

	data := strings.NewReader(`{"Firstname": "frederic", "Lastname": "ozanam","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`)
	req, err := inst.NewRequest("PUT", "/api/client", data)
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

	addclient(c, w, req)

	code := w.Code
	if code != http.StatusCreated {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body := w.Body.Bytes()
	expected := []byte(`{"Firstname":"frederic","Lastname":"ozanam","Address":"123 Easy St","Apt":"9","CrossStreet":"Main","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"UNK","ReferredBy":"districtofc","Notes":"landlord blahblahblah","Adultmales":"","Adultfemales":"","Fammbrs":null,"Financials":{"FatherIncome":"","MotherIncome":"","AFDCIncome":"","GAIncome":"","SSIIncome":"","UnemploymentInsIncome":"","SocialSecurityIncome":"","AlimonyIncome":"","ChildSupportIncome":"","Other1Income":"","Other1IncomeType":"","Other2Income":"","Other2IncomeType":"","Other3Income":"","Other3IncomeType":"","RentExpense":"","Section8Voucher":false,"UtilitiesExpense":"","WaterExpense":"","PhoneExpense":"","FoodExpense":"","GasExpense":"","CarPaymentExpense":"","TVInternetExpense":"","GarbageExpense":"","Other1Expense":"","Other1ExpenseType":"","Other2Expense":"","Other2ExpenseType":"","Other3Expense":"","Other3ExpenseType":"","TotalExpense":"","TotalIncome":""}}`)
	if !bytes.Contains(body, expected) {
		t.Errorf("got body %v (%v), want %v", body, string(body),
			expected)
	}

	q := datastore.NewQuery("SVDPClient")
	clients := make([]client, 0, 10)
	keys, err := q.GetAll(c, &clients)
	if err != nil {
		t.Fatalf("error on GetAll: %v", err)
		return
	}
	if len(clients) != 1 {
		t.Errorf("got %v records in query, expected %v", len(clients), 1)
	}

	for _, k := range keys {
		var updates []update
		q = datastore.NewQuery("SVDPUpdate").Ancestor(k)
		_, err = q.GetAll(c, &updates)
		if err != nil {
			t.Fatalf("error on SVDPUpdate GetAll: %v", err)
			return
		}
		if len(updates) != 1 {
			t.Errorf("got %v updates, expected 1", len(updates))
		}
		if updates[0].User != "test@example.org" {
			t.Errorf("update user is %v but expected test@example.org", updates[0].User)
		}
	}
}

func TestAddClientNamesEmpty(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	type reqresp struct {
		data     io.Reader
		expected []byte
	}

	missing := []reqresp{
		{strings.NewReader(`{"Firstname": "", "Lastname": "","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"UNK","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`),
			[]byte(`Firstname is empty and cannot be`)},
		{strings.NewReader(`{"Firstname": "Hello", "Lastname": "","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"UNK","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`),
			[]byte(`Lastname is empty and cannot be`)},
		{strings.NewReader(`{"Firstname": "Hello", "Lastname": "        ","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"UNK","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`),
			[]byte(`Lastname is empty and cannot be`)},
		{strings.NewReader(`{"Firstname": "     ", "Lastname": "xxx","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"UNK","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`),
			[]byte(`Firstname is empty and cannot be`)},
	}

	for i, x := range missing {
		req, err := inst.NewRequest("PUT", "/api/client", x.data)
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

		addclient(c, w, req)

		code := w.Code
		if code != http.StatusBadRequest {
			t.Errorf("req %v: got code %v, want %v", i, code, http.StatusBadRequest)
		}
		body := w.Body.Bytes()
		if !bytes.Contains(body, x.expected) {
			t.Errorf("req %v: got body %v (%v), want %v (%v)", i, body, string(body),
				x.expected, string(x.expected))
		}

		q := datastore.NewQuery("SVDPClient")
		clients := make([]client, 0, 10)
		keys, err := q.GetAll(c, &clients)
		if err != nil {
			t.Fatalf("error on GetAll: %v", err)
			return
		}
		if len(keys) != 0 {
			t.Errorf("req %v: got %v records in query, expected %v", i, len(keys), 0)
		}
	}
}

func TestAddClientInvalidValue(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	type reqresp struct {
		data     io.Reader
		expected []byte
	}

	invalid := []reqresp{
		{strings.NewReader(`{"Firstname": "Suzanne", "Lastname": "Test","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"unknown","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`),
			[]byte(`Ethnicity must be one of `)},
		{strings.NewReader(`{"Firstname": "Hello", "Lastname": "Test","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"caucasian","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`),
			[]byte(`Ethnicity must be one of `)},
		{strings.NewReader(`{"Firstname": "Hello", "Lastname": "Test","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"hispanic","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`),
			[]byte(`Ethnicity must be one of `)},
		{strings.NewReader(`{"Firstname": "Suzanne", "Lastname": "xxx","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"other","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`),
			[]byte(`Ethnicity must be one of `)},
	}

	for i, x := range invalid {
		req, err := inst.NewRequest("PUT", "/api/client", x.data)
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

		addclient(c, w, req)

		code := w.Code
		if code != http.StatusBadRequest {
			t.Errorf("req %v: got code %v, want %v", i, code, http.StatusBadRequest)
		}
		body := w.Body.Bytes()
		if !bytes.Contains(body, x.expected) {
			t.Errorf("req %v: got body %v (%v), want %v (%v)", i, body, string(body),
				x.expected, string(x.expected))
		}

		q := datastore.NewQuery("SVDPClient")
		clients := make([]client, 0, 10)
		keys, err := q.GetAll(c, &clients)
		if err != nil {
			t.Fatalf("error on GetAll: %v", err)
			return
		}
		if len(keys) != 0 {
			t.Errorf("req %v: got %v records in query, expected %v", i, len(keys), 0)
		}
	}
}

func TestGetAllClients(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclients := []client{
		{Firstname: "frederic", Lastname: "ozanam", Ethnicity: "UNK"},
		{Firstname: "John", Lastname: "Doe", Ethnicity: "O"},
		{Firstname: "Jane", Lastname: "Doe", Ethnicity: "H"},
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
	addTestUser(c, "test@example.org", true)

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

func TestGetAllVisitsInRange(t *testing.T) {
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
				Visitdate: "2013-05-03"},
			{Vincentians: "Irene, Jim",
				Visitdate: "2013-04-03"},
		},
		{
			{Vincentians: "Eileen, Lynn",
				Visitdate: "2013-07-03"},
			{Vincentians: "Stu & Anne",
				Visitdate: "2013-06-03"},
		},
	}

	numvisits := 0
	for i, viz := range visits {
		for _, vst := range viz {
			data, err := json.Marshal(vst)
			if err != nil {
				t.Fatalf("Failed to marshal %v", visits[i])
			}
			req, err := inst.NewRequest("PUT", "/api/visit/"+
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
			numvisits++
		}
	}

	req, err := inst.NewRequest("GET", "/api/getvisitsinrange/", nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	aetest.Login(&user.User{Email: "test@example.org"}, req)
	w := httptest.NewRecorder()

	c := appengine.NewContext(req)
	getvisitsinrange(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	createdvisits := []visit{}
	err = json.Unmarshal(body, &createdvisits)
	if err != nil {
		t.Errorf("error unmarshaling response to getvisitsinrange %v\n", err)
	}
	if len(createdvisits) != numvisits {
		t.Errorf("got %v visits, want %v\n", len(createdvisits),
			numvisits)
	}
	for _, viz := range visits {
		for _, vst := range viz {
			found := false
			for _, createdvisit := range createdvisits {
				if reflect.DeepEqual(createdvisit, vst) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("unable to find %v in %v",
					vst, &createdvisits)
			}
		}
	}
	for i, createdvisit := range createdvisits {
		var expecteddate string
		expecteddate = "2013-0" + strconv.Itoa(7-i) + "-03"
		if createdvisit.Visitdate != expecteddate {
			t.Errorf("dates not sorted? expected date %v, found %v for i %v",
				expecteddate, createdvisit.Visitdate, i)
		}
	}
}

func TestGetAllVisits(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}
	cltid, err := addclienttodb(newclient, inst)
	log.Printf("TestAllVisits: got %v from addclienttodb\n", cltid)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	visits := []visit{
		{Vincentians: "Michael, Mary Margaret",
			Visitdate: "2013-04-03"},
		{Vincentians: "Irene, Jim",
			Visitdate: "2013-05-03"},
	}
	for i := 0; i < len(visits); i++ {
		data, err := json.Marshal(visits[i])
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
	}

	req, err := inst.NewRequest("GET", "/api/getallvisits/"+
		strconv.FormatInt(cltid, 10), nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	aetest.Login(&user.User{Email: "test@example.org"}, req)
	w := httptest.NewRecorder()

	c := appengine.NewContext(req)
	getallvisits(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	createdvisitrecs := []visitrec{}
	err = json.Unmarshal(body, &createdvisitrecs)
	if err != nil {
		t.Errorf("error unmarshaling response to getvisits %v\n", err)
	}
	if len(createdvisitrecs) != len(visits) {
		t.Errorf("got %v visits, want %v\n", len(createdvisitrecs),
			1)
	}
	for i := 0; i < len(visits); i++ {
		found := false
		for j := 0; j < len(createdvisitrecs); j++ {
			if reflect.DeepEqual(createdvisitrecs[j].Visit, visits[i]) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unable to find %v in %v",
				visits[i], &createdvisitrecs)
		}
	}
}
func TestUpdateClient(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam", Ethnicity: "W"}
	id, err := addclienttodb(newclient, inst)
	log.Printf("TestUpdateClient: got %v from addclienttodb\n", id)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	data := strings.NewReader(`{"Firstname": "Frederic", "Lastname": "Ozanam", "Ethnicity": "W"}`)
	req, err := inst.NewRequest("PUT", "/client/"+
		strconv.FormatInt(id, 10), data)
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

	editclient(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body := w.Body.Bytes()
	expected := []byte(`{"Firstname":"Frederic","Lastname":"Ozanam","Address":"","Apt":"","CrossStreet":"","DOB":"","Phonenum":"","Altphonenum":"","Altphonedesc":"","Ethnicity":"W","ReferredBy":"","Notes":"","Adultmales":"","Adultfemales":"","Fammbrs":null,"Financials":{"FatherIncome":"","MotherIncome":"","AFDCIncome":"","GAIncome":"","SSIIncome":"","UnemploymentInsIncome":"","SocialSecurityIncome":"","AlimonyIncome":"","ChildSupportIncome":"","Other1Income":"","Other1IncomeType":"","Other2Income":"","Other2IncomeType":"","Other3Income":"","Other3IncomeType":"","RentExpense":"","Section8Voucher":false,"UtilitiesExpense":"","WaterExpense":"","PhoneExpense":"","FoodExpense":"","GasExpense":"","CarPaymentExpense":"","TVInternetExpense":"","GarbageExpense":"","Other1Expense":"","Other1ExpenseType":"","Other2Expense":"","Other2ExpenseType":"","Other3Expense":"","Other3ExpenseType":"","TotalExpense":"","TotalIncome":""}}`)
	if !bytes.Contains(body, expected) {
		t.Errorf("got body %v (%v), want %v (%v)", body, string(body),
			expected, string(expected))
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
	expectedc.Ethnicity = "W"
	if !reflect.DeepEqual(&clt, expectedc) {
		t.Errorf("db record shows %v, want %v", clt, expectedc)
	}

	var updates []update
	q := datastore.NewQuery("SVDPUpdate").Ancestor(key).Order("When")
	_, err = q.GetAll(c, &updates)
	if err != nil {
		t.Fatalf("error on SVDPUpdate GetAll: %v", err)
		return
	}
	if len(updates) != 2 {
		t.Errorf("got %v updates, expected 2", len(updates))
	}
	if updates[1].User != "test@example.org" {
		t.Errorf("update user is %v but expected test@example.org", updates[1].User)
	}
}

func TestVisitRouter(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}
	id, err := addclienttodb(newclient, inst)
	log.Printf("TestVisitRouter: got %v from addclienttodb\n", id)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	data := strings.NewReader(`{"Vincentians": "Michael, Mary Margaret", "VisitDate": "2011-04-01", "Giftcardamt": "100", "Numfoodboxes": "2"}`)
	req, err := inst.NewRequest("PUT", "/api/visit/"+
		strconv.FormatInt(id, 10), data)
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
	type routerreq struct {
		path string
		data io.Reader
		resp int
	}

	reqs := []routerreq{
		{"/api/visit/" + strconv.FormatInt(id, 10),
			strings.NewReader(`{"Vincentians": "Michael, Mary Margaret", "VisitDate": "2011-04-03", "Giftcardamt": "100",
				 "Numfoodboxes": "2"}`),
			200},
		{"/api/visit/" + strconv.FormatInt(newrec.ClientId, 10) + "/" + strconv.FormatInt(newrec.Id, 10) + "/edit",
			strings.NewReader(`{"Vincentians": "Michael, Mary Margaret", "VisitDate": "2011-04-01", "Giftcardamt": "199",
				 "Numfoodboxes": "2", "Comment": "updated"}`),
			200},
		{"/visit/", nil, 400},
		{"/viit/12345", nil, 400},
	}
	for _, u := range reqs {

		req, err := inst.NewRequest("PUT", u.path, u.data)
		req.Header = map[string][]string{"Content-Type": {"application/json"}}
		if err != nil {
			t.Fatalf("Failed to create req: %v", err)
		}

		aetest.Login(&user.User{Email: "test@example.org"}, req)

		w := httptest.NewRecorder()
		c := appengine.NewContext(req)

		visitrouter(c, w, req)

		code := w.Code
		if code != u.resp {
			t.Errorf("got code %v, want %v", code, u.resp)
		}
	}
}

func TestAddVisit(t *testing.T) {
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

	data := strings.NewReader(`{"Vincentians": "Michael, Mary Margaret", "VisitDate": "2011-04-03", "Giftcardamt": "100", "Numfoodboxes": "2"}`)
	req, err := inst.NewRequest("PUT", "/api/visit/"+
		strconv.FormatInt(id, 10), data)
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
	/*
		expected := []byte(`{"Vincentians":"Michael, Mary Margaret","Visitdate":"2011-04-03","Giftcardamt":"100","Numfoodboxes":"2"}`)
		if !bytes.Contains(body, expected) {
			t.Errorf("got body %v (%v), want %v (%v)", body, string(body),
				expected, string(expected))
		}
	*/
	key := datastore.NewKey(c, "SVDPClientVisit", "", newrec.Id,
		datastore.NewKey(c, "SVDPClient", "", id, nil))
	vst := visit{}
	if err := datastore.Get(c, key, &vst); err != nil {
		t.Fatalf("error on Get: %v", err)
		return
	}
	expectedv := &visit{}
	expectedv.Vincentians = "Michael, Mary Margaret"
	expectedv.Visitdate = "2011-04-03"
	expectedv.Giftcardamt = "100"
	expectedv.Numfoodboxes = "2"
	if !reflect.DeepEqual(&vst, expectedv) {
		t.Errorf("db record shows %v, want %v", vst, expectedv)
	}
	var updates []update
	q := datastore.NewQuery("SVDPUpdate").Ancestor(key)
	ukeys, err := q.GetAll(c, &updates)
	if err != nil {
		t.Fatalf("error on SVDPUpdate GetAll: %v", err)
		return
	}
	if len(ukeys) != 1 {
		t.Errorf("got %v updates, expected 1", len(updates))
	}
	if updates[0].User != "test@example.org" {
		t.Errorf("update user is %v but expected test@example.org", updates[0].User)
	}
}

func TestAddVisitMissingData(t *testing.T) {
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

	type reqresp struct {
		data     io.Reader
		expected []byte
	}

	missing := []reqresp{
		{strings.NewReader(`{"Vincentians": "Michael, Mary Margaret", "VisitDate": "", "Giftcardamt": "100", "Numfoodboxes": "2"}`),
			[]byte(`Visitdate is empty and cannot be`)},
		{strings.NewReader(`{"Vincentians": "Michael, Mary Margaret", "VisitDate": "      ", "Giftcardamt": "100", "Numfoodboxes": "2"}`),
			[]byte(`Visitdate is empty and cannot be`)},
		{strings.NewReader(`{"Vincentians": "", "VisitDate": "2013-05-03", "Giftcardamt": "100", "Numfoodboxes": "2"}`),
			[]byte(`Vincentians is empty and cannot be`)},
		{strings.NewReader(`{"Vincentians": "       ", "VisitDate": "2013-05-03", "Giftcardamt": "100", "Numfoodboxes": "2"}`),
			[]byte(`Vincentians is empty and cannot be`)},
	}

	for i, x := range missing {
		req, err := inst.NewRequest("PUT", "/api/visit/"+
			strconv.FormatInt(id, 10), x.data)
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
		if code != http.StatusBadRequest {
			t.Errorf("req %v: got code %v, want %v", i, code, http.StatusBadRequest)
		}
		body := w.Body.Bytes()

		if !bytes.Contains(body, x.expected) {
			t.Errorf("req %v: got %v and expected %v", i, string(body), string(x.expected))
		}
	}
}

func TestEditVisit(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}
	cltid, err := addclienttodb(newclient, inst)
	log.Printf("TestEditVisit: got %v from addclienttodb\n", cltid)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	visits := []visit{
		{Vincentians: "Michael, Mary Margaret",
			Visitdate: "2013-04-03"},
		{Vincentians: "Irene, Jim",
			Visitdate: "2013-05-03"},
	}

	vstids := make([]int64, 2)

	for i, v := range visits {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Failed to marshal %v", v)
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
		vstids[i] = newrec.Id
	}

	visits[0].Comment = "id is " + strconv.FormatInt(vstids[0], 10)
	visits[1].Comment = "id is " + strconv.FormatInt(vstids[1], 10)
	for i, v := range visits {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Failed to marshal %v", v)
		}
		req, err := inst.NewRequest("PUT", "/api/visit/"+
			strconv.FormatInt(cltid, 10)+
			"/"+strconv.FormatInt(vstids[i], 10)+"/edit", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to create req: %v", err)
		}
		req.Header = map[string][]string{
			"Content-Type": {"application/json"},
		}
		aetest.Login(&user.User{Email: "test@example.org"}, req)

		w := httptest.NewRecorder()
		c := appengine.NewContext(req)

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
	}

	req, err := inst.NewRequest("GET", "/api/visit/"+
		strconv.FormatInt(cltid, 10), nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	aetest.Login(&user.User{Email: "test@example.org"}, req)
	w := httptest.NewRecorder()

	c := appengine.NewContext(req)
	getallvisits(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	createdvisitrecs := []visitrec{}
	err = json.Unmarshal(body, &createdvisitrecs)
	if err != nil {
		t.Errorf("error unmarshaling response to getvisits %v\n", err)
	}
	if len(createdvisitrecs) != len(visits) {
		t.Errorf("got %v visits, want %v\n", len(createdvisitrecs),
			1)
	}
	for i := 0; i < len(visits); i++ {
		found := false
		for j := 0; j < len(createdvisitrecs); j++ {
			if reflect.DeepEqual(createdvisitrecs[j].Visit, visits[i]) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unable to find %v in %v",
				visits[i], &createdvisitrecs)
		}
	}
	for _, vid := range vstids {
		var updates []update
		k := datastore.NewKey(c, "SVDPClientVisit", "", vid,
			datastore.NewKey(c, "SVDPClient", "", cltid, nil))
		q := datastore.NewQuery("SVDPUpdate").Ancestor(k).Order("When")
		ukeys, err := q.GetAll(c, &updates)
		if err != nil {
			t.Fatalf("error on SVDPUpdate GetAll: %v", err)
			return
		}
		if len(ukeys) != 2 {
			t.Errorf("got %v updates, expected 2", len(updates))
		}
		if updates[1].User != "test@example.org" {
			t.Errorf("update user is %v but expected test@example.org", updates[1].User)
		}
	}
}

func TestEditVisitMissingData(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}
	cltid, err := addclienttodb(newclient, inst)
	log.Printf("TestEditVisit: got %v from addclienttodb\n", cltid)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	visits := []visit{
		{Vincentians: "Irene, Jim",
			Visitdate: "2013-05-03"},
	}

	vstids := make([]int64, 2)

	for i, v := range visits {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Failed to marshal %v", v)
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
			t.Errorf("got code %v, want %v", code, http.StatusOK)
		}

		body := w.Body.Bytes()
		newrec := &visitrec{}
		err = json.Unmarshal(body, newrec)
		if err != nil {
			t.Errorf("unable to parse %v: %v", string(body), err)
		}
		vstids[i] = newrec.Id
	}

	visits[0].Comment = "id is " + strconv.FormatInt(vstids[0], 10)
	visits[0].Visitdate = ""
	for i, v := range visits {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Failed to marshal %v", v)
		}
		req, err := inst.NewRequest("PUT", "/api/visit/"+
			strconv.FormatInt(cltid, 10)+
			"/"+strconv.FormatInt(vstids[i], 10)+"/edit", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Failed to create req: %v", err)
		}
		req.Header = map[string][]string{
			"Content-Type": {"application/json"},
		}
		aetest.Login(&user.User{Email: "test@example.org"}, req)

		w := httptest.NewRecorder()
		c := appengine.NewContext(req)

		visitrouter(c, w, req)

		code := w.Code
		if code != http.StatusBadRequest {
			t.Errorf("got code %v, want %v", code, http.StatusBadRequest)
		}

		body := w.Body.Bytes()
		expected := []byte("Visitdate is empty and cannot be")
		if !bytes.Contains(body, expected) {
			t.Errorf("got %v, expected %v", string(body), string(expected))
		}
	}

	req, err := inst.NewRequest("GET", "/api/visit/"+
		strconv.FormatInt(cltid, 10), nil)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	aetest.Login(&user.User{Email: "test@example.org"}, req)
	w := httptest.NewRecorder()

	c := appengine.NewContext(req)
	getallvisits(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	createdvisitrecs := []visitrec{}
	err = json.Unmarshal(body, &createdvisitrecs)
	if err != nil {
		t.Errorf("error unmarshaling response to getvisits %v\n", err)
	}
	if createdvisitrecs[0].Visit.Visitdate != "2013-05-03" {
		t.Errorf("Visitdate is %v but wanted %v", createdvisitrecs[0].Visit.Visitdate, "2013-05-03")
	}
}

func TestUpdateInvalidData(t *testing.T) {
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

	type reqresp struct {
		data     io.Reader
		expected []byte
	}
	bogus := []reqresp{
		{strings.NewReader(`{"Firstname": "Frederic", "Lastname": "Ozanam", "DOB":"alphabet"}`),
			[]byte(`parsing time "alphabet" as "2006-01-02": cannot parse "alphabet" as "2006"`)},
		{strings.NewReader(`{"Firstname": "Frederic", "Lastname": "Ozanam", "DOB":"1985-01-01", "Ethnicity":"unknown"}`),
			[]byte("Ethnicity must be one of ")},
	}
	for i, x := range bogus {
		req, err := inst.NewRequest("PUT", "/client/"+strconv.FormatInt(id,
			10), x.data)
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

		editclient(c, w, req)

		code := w.Code
		if code != http.StatusBadRequest {
			t.Errorf("test %v: got code %v, want %v", i, code, http.StatusBadRequest)
		}
		body := w.Body.Bytes()
		if !bytes.Contains(body, x.expected) {
			t.Errorf("req %v: got body %v (%v), expected %v (%v)", i, body, string(body),
				x.expected, string(x.expected))
		}
	}
}

func TestUpdateMissingData(t *testing.T) {
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

	type reqresp struct {
		data     io.Reader
		expected []byte
	}

	missing := []reqresp{
		{strings.NewReader(`{"Firstname": "", "Lastname": "Ozanam"}`),
			[]byte("Firstname is empty and cannot be")},
		{strings.NewReader(`{"Firstname": "Frederic", "Lastname": ""}`),
			[]byte("Lastname is empty and cannot be")},
	}
	for i, x := range missing {
		req, err := inst.NewRequest("PUT", "/client/"+strconv.FormatInt(id,
			10), x.data)
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

		editclient(c, w, req)

		code := w.Code
		if code != http.StatusBadRequest {
			t.Errorf("req %v: got code %v, want %v", i, code, http.StatusBadRequest)
		}
		body := w.Body.Bytes()
		if !bytes.Contains(body, x.expected) {
			t.Errorf("req %v: got body %v (%v), expected %v (%v)", i, body, string(body),
				x.expected, string(x.expected))
		}
	}
}

func TestUpdateMissingId(t *testing.T) {
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

	data := strings.NewReader(`{"Firstname": "Frederic", "Lastname": 
		"Ozanam"}`)
	req, err := inst.NewRequest("PUT", "/client/", data)
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

	editclient(c, w, req)

	code := w.Code
	if code != http.StatusBadRequest {
		t.Errorf("got code %v, want %v", code, http.StatusBadRequest)
	}
}

func TestUpdateMalformedId(t *testing.T) {
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

	data := strings.NewReader(`{"Firstname": "Frederic", "Lastname": 
		"Ozanam"}`)
	req, err := inst.NewRequest("PUT", "/client/"+"bogus-id", data)
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

	editclient(c, w, req)

	code := w.Code
	if code != http.StatusBadRequest {
		t.Errorf("got code %v, want %v", code, http.StatusBadRequest)
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

	aetest.Login(&user.User{Email: "adduser@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)

	addTestUser(c, "adduser@example.org", true)
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

func TestAuthorization(t *testing.T) {
	c, err := aetest.NewContext(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}
	defer c.Close()

	err = os.Setenv("BOOTSTRAP_USER", "hello@example.org")
	if err != nil {
		t.Fatalf("failed to set BOOTSTRAP_USER: %v", err.Error())
	}

	auth, err := userauthorized(c, "hello@example.org")
	if err != nil {
		t.Fatalf("auth error: %v", err.Error())
	}
	if !auth {
		t.Errorf("auth failed for bootstrap user")
	}
	addTestUser(c, "frederic@example.org", true)

	auth, err = userauthorized(c, "frederic@example.org")

	if err != nil {
		t.Fatalf("auth error: %v", err.Error())
	}
	if !auth {
		t.Errorf("auth failed for frederic@example.org")
	}
	auth, err = userauthorized(c, "fred@example.org")

	if err != nil {
		t.Fatalf("auth error: %v", err.Error())
	}
	if auth {
		t.Errorf("auth worked and shouldn't have for fred@example.org")
	}
}

func TestAddUsers(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	data := strings.NewReader(`{"Ids": [0, 0], "Aus": [{"Email": "fred1@example.org"}, {"Email": "fred2@example.org"}], "DeletedIds": []}`)
	req, err := inst.NewRequest("PUT", "/editusers", data)
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

	editusers(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body := w.Body.Bytes()
	c.Infof("got response %v", string(body))

	var b2 useredit
	err = json.Unmarshal(body, &b2)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}
	e := [2]string{"fred1@example.org", "fred2@example.org"}
	for i := 0; i < len(e); i++ {
		a, err := userauthorized(c, e[i])
		if err != nil {
			t.Fatalf("authorization error: %v", err)
		}
		if !a {
			t.Errorf("user %v not authorized", e[i])
		}
	}
	var updates []update
	q := datastore.NewQuery("SVDPUserUpdate")
	ukeys, err := q.GetAll(c, &updates)
	if err != nil {
		t.Fatalf("error on SVDPUserUpdate GetAll: %v", err)
		return
	}
	if len(ukeys) != 1 {
		t.Errorf("got %v updates, expected 1", len(updates))
	}
	if updates[0].User != "test@example.org" {
		t.Errorf("update user is %v but expected test@example.org", updates[0].User)
	}
}

func TestAddUsersMissingIds(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	data := strings.NewReader(`{"Ids": [], "Aus": [{"Email": "fred1@example.org"}, {"Email": "fred2@example.org"}], "DeletedIds": []}`)
	req, err := inst.NewRequest("PUT", "/editusers", data)
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

	editusers(c, w, req)

	code := w.Code
	if code != http.StatusBadRequest {
		t.Errorf("got code %v, want %v", code, http.StatusBadRequest)
	}
	body := w.Body.Bytes()
	c.Infof("got response %v", string(body))

}

func TestEditUsers(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	data := strings.NewReader(`{"Ids": [0, 0], "Aus": [{"Email": "fred1@example.org"}, {"Email": "fred2@example.org"}], "DeletedIds": []}`)
	req, err := inst.NewRequest("PUT", "/editusers", data)
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

	editusers(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body := w.Body.Bytes()
	c.Infof("got response %v", string(body))

	var b2 useredit
	err = json.Unmarshal(body, &b2)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}
	b2.Aus[0].Email = "fred3@example.org"
	b2.Aus[1].Email = "fred4@example.org"

	data1, err := json.Marshal(&b2)
	req1, err := inst.NewRequest("PUT", "/editusers", bytes.NewReader(data1))
	if err != nil {
		t.Fatalf("Failed to create req1: %v", err)
	}
	req1.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req1)

	w1 := httptest.NewRecorder()

	editusers(c, w1, req1)

	code = w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body1 := w.Body.Bytes()
	c.Infof("got response %v", string(body1))

	var b3 useredit
	err = json.Unmarshal(body, &b3)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}
	e1 := [2]string{"fred3@example.org", "fred4@example.org"}
	for i := 0; i < len(e1); i++ {
		a, err := userauthorized(c, e1[i])
		if err != nil {
			t.Fatalf("authorization error: %v", err)
		}
		if !a {
			t.Errorf("user %v not authorized", e1[i])
		}
	}
	var updates []update
	q := datastore.NewQuery("SVDPUserUpdate")
	ukeys, err := q.GetAll(c, &updates)
	if err != nil {
		t.Fatalf("error on SVDPUserUpdate GetAll: %v", err)
		return
	}
	if len(ukeys) != 2 {
		t.Errorf("got %v updates, expected 2", len(updates))
	}
	if updates[0].User != "test@example.org" {
		t.Errorf("update user is %v but expected test@example.org", updates[0].User)
	}
}

func TestEditUsersNotAdmin(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	data := strings.NewReader(`{"Ids": [0, 0], "Aus": [{"Email": "fred1@example.org"}, {"Email": "fred2@example.org"}], "DeletedIds": []}`)
	req, err := inst.NewRequest("PUT", "/editusers", data)
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	c := appengine.NewContext(req)
	addTestUser(c, "test@example.org", false)

	editusers(c, w, req)

	code := w.Code
	if code != http.StatusForbidden {
		t.Errorf("got code %v, want %v", code, http.StatusForbidden)
	}
}

func TestDeleteUsers(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	data := strings.NewReader(`{"Ids": [0, 0], "Aus": [{"Email": "fred1@example.org"}, {"Email": "fred2@example.org"}], "DeletedIds": []}`)
	req, err := inst.NewRequest("PUT", "/editusers", data)
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

	editusers(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body := w.Body.Bytes()
	c.Infof("got response %v", string(body))

	var b2 useredit
	err = json.Unmarshal(body, &b2)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}
	id0 := b2.Ids[0]
	id1 := b2.Ids[1]

	b2.Aus = make([]appuser, 1)
	b2.Ids = make([]int64, 1)
	b2.DeletedIds = make([]int64, 1)

	b2.Aus[0].Email = "fred4@example.org"
	b2.Ids[0] = id1
	b2.DeletedIds[0] = id0

	data1, err := json.Marshal(&b2)
	req1, err := inst.NewRequest("PUT", "/editusers", bytes.NewReader(data1))
	if err != nil {
		t.Fatalf("Failed to create req1: %v", err)
	}
	req1.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req1)

	w1 := httptest.NewRecorder()

	editusers(c, w1, req1)

	code = w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body1 := w.Body.Bytes()
	c.Infof("got response %v", string(body1))

	var b3 useredit
	err = json.Unmarshal(body, &b3)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}
	a, err := userauthorized(c, "fred4@example.org")
	if err != nil {
		t.Fatalf("authorization error: %v", err)
	}
	if !a {
		t.Errorf("user %v not authorized", "fred4@example.org")
	}
	a, err = userauthorized(c, "fred3@example.org")
	if err != nil {
		t.Fatalf("authorization error: %v", err)
	}
	if a {
		t.Errorf("user %v is authorized and should not be", "fred3@example.org")
	}
}

func TestGetUsers(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	data := strings.NewReader(`{"Ids": [0, 0], "Aus": [{"Email": "fred1@example.org"}, {"Email": "fred2@example.org"}], "DeletedIds": []}`)
	req, err := inst.NewRequest("PUT", "/editusers", data)
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

	editusers(c, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body := w.Body.Bytes()
	c.Infof("got response %v", string(body))

	req1, err := inst.NewRequest("GET", "/api/clients", nil)
	if err != nil {
		t.Fatalf("Failed to create req1: %v", err)
	}
	req1.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req1)

	w1 := httptest.NewRecorder()

	getallusers(c, w1, req1)

	code = w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusCreated)
	}
	body1 := w.Body.Bytes()
	c.Infof("got response %v", string(body1))

	var b3 useredit
	err = json.Unmarshal(body, &b3)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}
	e1 := [2]string{"fred1@example.org", "fred2@example.org"}
	for i := 0; i < len(e1); i++ {
		if e1[i] != b3.Aus[i].Email {
			t.Errorf("user %v not in user list", e1[i])
		}
	}
}

func addTestUser(c appengine.Context, u string, admin bool) (*datastore.Key, error) {
	q := datastore.NewQuery("SVDPUser").KeysOnly().Filter("Email =", u).Filter("IsAdmin =", admin)
	keys, err := q.GetAll(c, nil)
	if err != nil {
		c.Errorf("Failed to query for users: %v", err)
		return nil, err
	}

	if len(keys) > 0 {
		c.Infof("user %v/%v already in datastore", u, admin)
		return keys[0], nil
	}

	newuser := &appuser{Email: u, IsAdmin: admin}

	id, err := datastore.Put(c, datastore.NewIncompleteKey(c, "SVDPUser",
		nil), newuser)

	c.Infof("id=%v, appuser=%v, err=%v", id, newuser, err)
	if err != nil {
		c.Errorf("Failed to put user: %v", err)
		return nil, err
	}
	return id, nil
}
