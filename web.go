package frederic

//TODO: -common web page, api auth logic
//      -figure out testing of update pages

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type ContextHandler struct {
	Real func(appengine.Context, http.ResponseWriter, *http.Request)
}

type client struct {
	Firstname   string
	Lastname    string
	Address     string
	Apt         string
	DOB         string
	Phonenum    string
	Addlmales   string
	Addlfemales string
	Fammbrs     []fammbr
	Financials  financials
}

type fammbr struct {
	Name   string
	DOB    string
	Female bool
}

type financials struct {
	FatherIncome          string
	MotherIncome          string
	AFDCIncome            string
	GAIncome              string
	SSIIncome             string
	UnemploymentInsIncome string
	SocialSecurityIncome  string
	AlimonyIncome         string
	ChildSupportIncome    string
	Other1Income          string
	Other1IncomeType      string
	Other2Income          string
	Other2IncomeType      string
	Other3Income          string
	Other3IncomeType      string
	RentExpense           string
	UtilitiesExpense      string
	WaterExpense          string
	PhoneExpense          string
	FoodExpense           string
	GasExpense            string
	CarPaymentExpense     string
	TVInternetExpense     string
	GarbageExpense        string
	Other1Expense         string
	Other1ExpenseType     string
	Other2Expense         string
	Other2ExpenseType     string
	Other3Expense         string
	Other3ExpenseType     string
	TotalExpense          string
	TotalIncome           string
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
	http.Handle("/", ContextHandler{homepage})
	http.Handle("/clients", ContextHandler{listclientspage})
	http.Handle("/client", ContextHandler{getclientpage})
	http.Handle("/editclient", ContextHandler{editclientpage})
	http.Handle("/addclient", ContextHandler{newclientpage})
	http.Handle("/api/client", ContextHandler{addclient})
	http.Handle("/api/client/", ContextHandler{editclient})
	http.Handle("/api/getallclients", ContextHandler{getallclients})
}

var funcMap = template.FuncMap{"age": age}
var templates = template.Must(template.New("client").Funcs(funcMap).ParseGlob("*.html"))

func age(dobs string) string {
	if len(dobs) == 0 {
		return ""
	}
	dob, err := time.Parse("2006-01-02", dobs)
	if err != nil {
		return ""
	}
	agens := time.Since(dob)
	return strconv.FormatFloat(agens.Hours()/float64(24*365), 'f', 0, 64)
}

func homepage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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
	err := templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func listclientspage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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

	clientrecs := make([]clientrec, len(keys))
	for i := 0; i < len(clients); i++ {
		clientrecs[i].Clt = clients[i]
		clientrecs[i].Id = keys[i].IntID()
	}
	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		Clients      []clientrec
	}{
		u.Email,
		l,
		clientrecs,
	}
	err = templates.ExecuteTemplate(w, "clients.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getclientpage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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
		Clientrec    clientrec
	}{
		u.Email,
		l,
		clientrec{id, clt},
	}

	err = templates.ExecuteTemplate(w, "client.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newclientpage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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

	clt := client{}

	data := struct {
		U, LogoutUrl string
		Clientrec    clientrec
	}{
		u.Email,
		l,
		clientrec{0, clt},
	}
	err := templates.ExecuteTemplate(w, "newclient.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func editclientpage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
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
		Clientrec    clientrec
	}{
		u.Email,
		l,
		clientrec{id, clt},
	}
	err = templates.ExecuteTemplate(w, "editclient.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
