package frederic

//TODO: -common web page, api auth logic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

func addclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	new := &client{}
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	err = json.Unmarshal(body, new)
	c.Infof("api/addclient: got %v\n", string(body))
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
			new.DOB, new.Phonenum, new.Fammbrs},
	}
	b, err := json.Marshal(newrec)
	c.Infof("returning %v\n", string(b))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(b))
}

func editclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	u := user.Current(c)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cltrec := &clientrec{}
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	err = json.Unmarshal(body, cltrec)
	c.Infof("api/editclient: got %v\n", string(body))
	if err != nil {
		c.Errorf("unmarshaling error:%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(cltrec.Clt.DOB) > 0 {
		if _, err = time.Parse("2006-01-02", cltrec.Clt.DOB); err != nil {
			c.Errorf("unable to parse DOB %v, err %v",
				cltrec.Clt.DOB, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	clt := &client{cltrec.Clt.Firstname, cltrec.Clt.Lastname,
		cltrec.Clt.Address, cltrec.Clt.Apt, cltrec.Clt.DOB,
		cltrec.Clt.Phonenum, cltrec.Clt.Fammbrs}
	ikey := datastore.NewKey(c, "SVDPClient", "", cltrec.Id, nil)
	key, err := datastore.Put(c, ikey, clt)
	if err != nil {
		c.Errorf("datastore error on Put: :%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newrec := &clientrec{key.IntID(),
		client{cltrec.Clt.Firstname, cltrec.Clt.Lastname,
			cltrec.Clt.Address, cltrec.Clt.Apt, cltrec.Clt.DOB,
			cltrec.Clt.Phonenum, cltrec.Clt.Fammbrs},
	}
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
	u := user.Current(c)
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
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
			clients[i].Fammbrs}}
	}
	c.Debugf("getallclients: clientrecs = %v\n", clientrecs)
	b, err := json.Marshal(clientrecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(b))
}
