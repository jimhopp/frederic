package frederic

//TODO: -common web page, api auth logic
//      -figure out testing of update pages

import (
	"html/template"
	"net/http"
	"strconv"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type ContextHandler struct {
	Real func(appengine.Context, http.ResponseWriter, *http.Request)
}

type client struct {
	Firstname string
	Lastname  string
}

type clientrec struct {
	Id  int64
	Clt client
}

func (f ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	f.Real(c, w, r)
}

func init() {
	http.Handle("/", ContextHandler{home})
	http.Handle("/clients", ContextHandler{listclients})
	http.Handle("/client", ContextHandler{getclient})
	http.Handle("/addclient", ContextHandler{newclient})
	http.Handle("/api/addclient", ContextHandler{addclient})
	http.Handle("/api/getallclients", ContextHandler{getallclients})
}

func home(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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
	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")
	data := struct {
		U, LogoutUrl string
	}{
		u.Email,
		l,
	}
	err := homepageTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var homepageTemplate = template.Must(template.ParseFiles("home.html",
	"scripts.html", "header.html"))

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
	clients := make([]client, 0, 10)
	keys, err := q.GetAll(c, &clients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		Clients      []client
		Keys         []*datastore.Key
	}{
		u.Email,
		l,
		clients,
		keys,
	}
	err = clientsTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var clientsTemplate = template.Must(template.ParseFiles("clients.html",
	"scripts.html", "header.html"))

func getclient(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query()
	ids, ok := query["id"]
	if !ok {
		http.Error(w, "id parm missing or mis-formed",
			http.StatusNotFound)
		return
	}
	id, err := strconv.ParseInt(ids[0], 10, 64)
	if err != nil {
		c.Warningf("got error %v trying to parse id %v\n", err, id)
		http.Error(w, "unable to find client", http.StatusNotFound)
		return
	}
	key := datastore.NewKey(c, "SVDPClient", "", id, nil)
	var clt client
	err = datastore.Get(c, key, &clt)
	if err != nil {
		c.Warningf("got error %v on datastore get for key %v\n", err,
			key)
		http.Error(w, "unable to find client",
			http.StatusNotFound)
		return
	}

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		Client       client
		Key          int64
	}{
		u.Email,
		l,
		clt,
		key.IntID(),
	}
	err = clientTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var clientTemplate = template.Must(template.ParseFiles("client.html",
	"scripts.html", "header.html"))

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

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
	}{
		u.Email,
		l,
	}
	err := newClientTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var newClientTemplate = template.Must(template.ParseFiles("newclient.html",
	"scripts.html", "header.html"))
