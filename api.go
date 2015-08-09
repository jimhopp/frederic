package frederic

//TODO: -common web page, api auth logic

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		client{new.Firstname, new.Lastname},
	}
	b, err := json.Marshal(newrec)
	c.Infof("returning %v\n", string(b))
	w.WriteHeader(http.StatusCreated)
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
			clients[i].Lastname}}
	}
	c.Debugf("getallclients: clientrecs = %v\n", clientrecs)
	b, err := json.Marshal(clientrecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(b))
}
