package frederic
//TODO: -common web page, api auth logic
//      -template for header on pages
//      -figure out testing of update pages
//      -rename files

import (
    "net/http"
    "html/template"

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
    http.Handle("/clients", ContextHandler{listclients})
    http.Handle("/addclient", ContextHandler{newclient})
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
    type homepage struct{U, Logouturl string}
    l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")
    h := homepage{u.Email, l} 
    err := homepageTemplate.Execute(w, h)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

var homepageTemplate = template.Must(template.ParseFiles("home.html", "scripts.html"))

func listclients(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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
 
    q := datastore.NewQuery("SVDPClient")
    Clients := make([]client, 0, 10)
    if _, err := q.GetAll(c, &Clients); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err := clientsTemplate.Execute(w, Clients)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
var clientsTemplate = template.Must(template.ParseFiles("clients.html", 
    "scripts.html"))

func newclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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
 
    err := newClientTemplate.Execute(w, u)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
var newClientTemplate = template.Must(template.ParseFiles("newclient.html", "scripts.html"))

