package frederic

import (
    "fmt"
    "net/http"

    "appengine"
    "appengine/user"
)

type ContextHandler struct {
    Real func(appengine.Context, http.ResponseWriter, *http.Request)
}

func (f ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    f.Real(c, w, r)
}

func init() {
    http.Handle("/", ContextHandler{handler})
    http.Handle("/api/addclient", ContextHandler{addclient})
    http.Handle("/api/getclient", ContextHandler{getclient})
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
    w.WriteHeader(http.StatusCreated)
    fmt.Fprintln(w, "client added")
}

func getclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
    u:= user.Current(c)
    if u == nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    w.WriteHeader(http.StatusCreated)
    fmt.Fprintln(w, "client view")
}
