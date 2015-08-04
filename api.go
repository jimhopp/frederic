package frederic
//TODO: -common web page, api auth logic

import (
    "fmt"
    "net/http"
    "encoding/json"
    "log"

    "appengine"
    "appengine/user"
    "appengine/datastore"
)

func addclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
    u:= user.Current(c)
    if u == nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    new := &client{}
    body := make([]byte, r.ContentLength)
    _, err := r.Body.Read(body)
    err = json.Unmarshal(body, new)
    log.Printf("api/addclient: got %v\n", string(body))
    if err != nil {
	log.Printf("unmarshaling error:%v\n", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    key := datastore.NewIncompleteKey(c, "SVDPClient", nil)
    _, err = datastore.Put(c, key, new)
    if err != nil {
         http.Error(w, err.Error(), http.StatusInternalServerError)
         return
    }

    b, err := json.Marshal(new)
    w.WriteHeader(http.StatusCreated)
    fmt.Fprint(w, string(b))
}

func getallclients(c appengine.Context, w http.ResponseWriter, r *http.Request) {
    u:= user.Current(c)
    if u == nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    q := datastore.NewQuery("SVDPClient")
    clients := make([]client, 0, 10)
    if _, err := q.GetAll(c, &clients); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)

    b, err := json.Marshal(clients)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    fmt.Fprint(w,string(b))
}
