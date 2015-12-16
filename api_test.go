package frederic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

	data := strings.NewReader(`{"Firstname": "frederic", "Lastname": "ozanam","Address":"123 Easy St","CrossStreet":"Main","Apt":"9","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"unknown","ReferredBy":"districtofc","Notes":"landlord blahblahblah"}`)
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
	expected := []byte(`{"Firstname":"frederic","Lastname":"ozanam","Address":"123 Easy St","Apt":"9","CrossStreet":"Main","DOB":"1823-04-13","Phonenum":"650-555-1212","Altphonenum":"650-767-2676","Altphonedesc":"POP-CORN","Ethnicity":"unknown","ReferredBy":"districtofc","Notes":"landlord blahblahblah","Adultmales":"","Adultfemales":"","Fammbrs":null,"Financials":{"FatherIncome":"","MotherIncome":"","AFDCIncome":"","GAIncome":"","SSIIncome":"","UnemploymentInsIncome":"","SocialSecurityIncome":"","AlimonyIncome":"","ChildSupportIncome":"","Other1Income":"","Other1IncomeType":"","Other2Income":"","Other2IncomeType":"","Other3Income":"","Other3IncomeType":"","RentExpense":"","Section8Voucher":false,"UtilitiesExpense":"","WaterExpense":"","PhoneExpense":"","FoodExpense":"","GasExpense":"","CarPaymentExpense":"","TVInternetExpense":"","GarbageExpense":"","Other1Expense":"","Other1ExpenseType":"","Other2Expense":"","Other2ExpenseType":"","Other3Expense":"","Other3ExpenseType":"","TotalExpense":"","TotalIncome":""}}`)
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
		req, err := inst.NewRequest("PUT", "/visit/"+
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

	newclient := client{Firstname: "frederic", Lastname: "ozanam"}
	id, err := addclienttodb(newclient, inst)
	log.Printf("TestUpdateClient: got %v from addclienttodb\n", id)
	if err != nil {
		t.Fatalf("unable to add client: %v", err)
	}

	data := strings.NewReader(`{"Firstname": "Frederic", "Lastname": "Ozanam"}`)
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
	expected := []byte(`{"Firstname":"Frederic","Lastname":"Ozanam","Address":"","Apt":"","CrossStreet":"","DOB":"","Phonenum":"","Altphonenum":"","Altphonedesc":"","Ethnicity":"","ReferredBy":"","Notes":"","Adultmales":"","Adultfemales":"","Fammbrs":null,"Financials":{"FatherIncome":"","MotherIncome":"","AFDCIncome":"","GAIncome":"","SSIIncome":"","UnemploymentInsIncome":"","SocialSecurityIncome":"","AlimonyIncome":"","ChildSupportIncome":"","Other1Income":"","Other1IncomeType":"","Other2Income":"","Other2IncomeType":"","Other3Income":"","Other3IncomeType":"","RentExpense":"","Section8Voucher":false,"UtilitiesExpense":"","WaterExpense":"","PhoneExpense":"","FoodExpense":"","GasExpense":"","CarPaymentExpense":"","TVInternetExpense":"","GarbageExpense":"","Other1Expense":"","Other1ExpenseType":"","Other2Expense":"","Other2ExpenseType":"","Other3Expense":"","Other3ExpenseType":"","TotalExpense":"","TotalIncome":""}}`)
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
	if !reflect.DeepEqual(&clt, expectedc) {
		t.Errorf("db record shows %v, want %v", clt, expectedc)
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
	req, err := inst.NewRequest("PUT", "/visit/"+
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

	data := strings.NewReader(`{"Firstname": "Frederic", "Lastname": 
		"Ozanam", "DOB":"alphabet"}`)
	req, err := inst.NewRequest("PUT", "/client/"+strconv.FormatInt(id,
		10), data)
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
