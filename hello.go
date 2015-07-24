package frederic

import (
    "fmt"
    "net/http"
    "encoding/json"

    "appengine"
    "appengine/user"
    "appengine/datastore"
)

type ContextHandler struct {
    Real func(appengine.Context, http.ResponseWriter, *http.Request)
}

type client struct {
    Firstname string
    Lastname  string
}

func (f ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    f.Real(c, w, r)
}

func init() {
    http.Handle("/", ContextHandler{handler})
    http.Handle("/api/addclient", ContextHandler{addclient})
    http.Handle("/api/getallclients", ContextHandler{getallclients})
}

func handler(c appengine.Context, w http.ResponseWriter, r *http.Request) {
    u := user.Current(c)
    if u == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
    fmt.Fprintf(w, "This is the SVdP Clients homepage.\n\nYou are authenticated as %v", u)
}

func addclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
    u:= user.Current(c)
    if u == nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    new := &client{}
    str := `{"Firstname": "frederic", "Lastname": "ozanam"}`
    json.Unmarshal([]byte(str), new)

    key := datastore.NewIncompleteKey(c, "SVDPClient", nil)
    _, err := datastore.Put(c, key, new)
    if err != nil {
         http.Error(w, err.Error(), http.StatusInternalServerError)
         return
    }

    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, "client %v, %v added", new.Lastname, new.Firstname)
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
    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, "Total clients: %v\n", len(clients))
    for i := 0; i < len(clients); i++ {
        fmt.Fprintf(w, "%v: %v, %v\n", i, clients[i].Lastname,
            clients[i].Firstname)
    }
}
