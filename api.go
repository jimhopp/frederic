package frederic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

func apiuserOK(c appengine.Context, w http.ResponseWriter) bool {
	if !userauthenticated(c) {
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	if ok, _ := userauthorized(c, user.Current(c).Email); !ok {
		w.WriteHeader(http.StatusForbidden)
		return false
	}
	return true
}

func addclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !apiuserOK(c, w) {
		return
	}

	new := &client{}
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	err = json.Unmarshal(body, new)
	c.Infof("addclient: got %v\n", string(body))
	if err != nil {
		c.Errorf("unmarshaling error:%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ikey := datastore.NewIncompleteKey(c, "SVDPClient", nil)
	key, err := datastore.Put(c, ikey, new)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newrec := &clientrec{key.IntID(),
		client{new.Firstname, new.Lastname, new.Address, new.Apt,
			new.DOB, new.Phonenum, new.Addlmales, new.Addlfemales,
			new.Fammbrs, new.Financials},
	}
	b, err := json.Marshal(newrec)
	c.Infof("returning %v\n", string(b))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(b))
}

func editclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !apiuserOK(c, w) {
		return
	}

	clt := &client{}
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	err = json.Unmarshal(body, clt)
	c.Infof("api/editclient: got %v\n", string(body))
	if err != nil {
		c.Errorf("unmarshaling error:%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	re, err := regexp.Compile("[0-9]+")
	idstr := re.FindString(r.URL.Path)
	c.Debugf("parsed id %v from %v", idstr, r.URL.Path)

	if idstr == "" {
		c.Errorf("id is missing for update request: path %v, data %v",
			r.URL.Path, string(body))
		http.Error(w,
			fmt.Sprintf("id is missing in path for update request %v", string(body)),
			http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		c.Errorf("unable to parse id %v as int64: %v", id, err.Error())
		http.Error(w,
			fmt.Sprintf("unable to parse id %v as int64: %v", id,
				err.Error()),
			http.StatusBadRequest)
		return
	}

	if len(clt.DOB) > 0 {
		if _, err = time.Parse("2006-01-02", clt.DOB); err != nil {
			c.Errorf("unable to parse DOB %v, err %v",
				clt.DOB, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	ikey := datastore.NewKey(c, "SVDPClient", "", id, nil)
	key, err := datastore.Put(c, ikey, clt)
	if err != nil {
		c.Errorf("datastore error on Put: :%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newrec := &clientrec{key.IntID(), *clt}

	b, err := json.Marshal(newrec)
	if err != nil {
		c.Errorf("marshaling error:%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Infof("returning %v\n", string(b))
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(b))
}

func addvisit(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !apiuserOK(c, w) {
		return
	}

	vst := &visit{}
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	err = json.Unmarshal(body, vst)
	c.Infof("api/addvisit: got %v\n", string(body))
	if err != nil {
		c.Errorf("unmarshaling error:%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	re, err := regexp.Compile("[0-9]+")
	idstr := re.FindString(r.URL.Path)
	c.Debugf("parsed id %v from %v", idstr, r.URL.Path)

	if idstr == "" {
		c.Errorf("id is missing for add visit request: path %v, data %v",
			r.URL.Path, string(body))
		http.Error(w,
			fmt.Sprintf("id is missing in path for add visit request %v", string(body)),
			http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		c.Errorf("unable to parse id %v as int64: %v", id, err.Error())
		http.Error(w,
			fmt.Sprintf("unable to parse id %v as int64: %v", id,
				err.Error()),
			http.StatusBadRequest)
		return
	}

	if len(vst.Visitdate) > 0 {
		if _, err = time.Parse("2006-01-02", vst.Visitdate); err != nil {
			c.Errorf("unable to parse visit date %v, err %v",
				vst.Visitdate, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	ikey := datastore.NewIncompleteKey(c, "SVDPClientVisit",
		datastore.NewKey(c, "SVDPClient", "", id, nil))
	key, err := datastore.Put(c, ikey, vst)
	if err != nil {
		c.Errorf("datastore error on Put: :%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newrec := &visitrec{key.IntID(), id, *vst}

	b, err := json.Marshal(newrec)
	if err != nil {
		c.Errorf("marshaling error:%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Infof("returning %v\n", string(b))
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(b))
}

func getallclients(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !apiuserOK(c, w) {
		return
	}

	q := datastore.NewQuery("SVDPClient")
	clients := make([]client, 0, 10)
	ids, err := q.GetAll(c, &clients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Debugf("getallclients: got keys %v\n", ids)
	w.WriteHeader(http.StatusOK)

	clientrecs := make([]clientrec, len(clients))
	for i := 0; i < len(clients); i++ {
		clientrecs[i] = clientrec{ids[i].IntID(), client{clients[i].Firstname,
			clients[i].Lastname, clients[i].Address,
			clients[i].Apt, clients[i].DOB, clients[i].Phonenum,
			clients[i].Addlmales, clients[i].Addlfemales,
			clients[i].Fammbrs, clients[i].Financials}}
	}
	c.Debugf("getallclients: clientrecs = %v\n", clientrecs)
	b, err := json.Marshal(clientrecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(b))
}

func getallvisits(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !apiuserOK(c, w) {
		return
	}

	re, err := regexp.Compile("[0-9]+")
	idstr := re.FindString(r.URL.Path)
	c.Debugf("parsed id %v from %v", idstr, r.URL.Path)

	if idstr == "" {
		c.Errorf("id is missing for request: path %v", r.URL.Path)
		http.Error(w,
			fmt.Sprintf("id is missing in path %v", r.URL.Path),
			http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		c.Errorf("unable to parse id %v as int64: %v", id, err.Error())
		http.Error(w,
			fmt.Sprintf("unable to parse id %v as int64: %v", id,
				err.Error()),
			http.StatusBadRequest)
		return
	}
	q := datastore.NewQuery("SVDPClientVisit").Ancestor(datastore.NewKey(
		c, "SVDPClient", "", id, nil))
	visits := make([]visit, 0, 10)
	ids, err := q.GetAll(c, &visits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Debugf("getallvisits: got keys %v\n", ids)
	w.WriteHeader(http.StatusOK)

	visitrecs := make([]visitrec, len(visits))
	for i := 0; i < len(visits); i++ {
		visitrecs[i] = visitrec{ids[i].IntID(), id,
			visit{visits[i].Vincentians, visits[i].Visitdate,
				visits[i].Assistancerequested, visits[i].Giftcardamt,
				visits[i].Numfoodboxes, visits[i].Rentassistance,
				visits[i].Utilitiesassistance,
				visits[i].Waterbillassistance,
				visits[i].Otherassistancetype,
				visits[i].Otherassistanceamt,
				visits[i].Vouchersclothing, visits[i].Vouchersfurniture,
				visits[i].Vouchersother, visits[i].Comment}}
	}
	c.Debugf("getallclients: visitrecs = %v\n", visitrecs)
	b, err := json.Marshal(visitrecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(b))
}
