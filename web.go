package frederic

//TODO: -figure out testing of update pages

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
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
	Firstname    string
	Lastname     string
	Address      string
	Apt          string
	CrossStreet  string
	DOB          string
	Phonenum     string
	Altphonenum  string
	Altphonedesc string
	Ethnicity    string
	ReferredBy   string
	Notes        string
	Adultmales   string
	Adultfemales string
	Fammbrs      []fammbr
	Financials   financials
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
	Section8Voucher       bool
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

type visit struct {
	Vincentians         string
	Visitdate           string
	Assistancerequested string
	Giftcardamt         string
	Numfoodboxes        string
	Rentassistance      string
	Utilitiesassistance string
	Waterbillassistance string
	Otherassistancetype string
	Otherassistanceamt  string
	Vouchersclothing    string
	Vouchersfurniture   string
	Vouchersother       string
	Comment             string
}

type clientrec struct {
	Id  int64
	Clt client
}

type visitrec struct {
	Id       int64
	ClientId int64
	Visit    visit
}

func (f ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	f.Real(c, w, r)
}

func init() {
	http.Handle("/", ContextHandler{homepage})
	http.Handle("/clients", ContextHandler{listclientspage})
	http.Handle("/client/", ContextHandler{getclientpage})
	http.Handle("/editclient/", ContextHandler{editclientpage})
	http.Handle("/addclient", ContextHandler{newclientpage})
	http.Handle("/recordvisit/", ContextHandler{recordvisitpage})
	http.Handle("/users", ContextHandler{edituserspage})
	http.Handle("/api/client", ContextHandler{addclient})
	http.Handle("/api/client/", ContextHandler{editclient})
	http.Handle("/api/visit/", ContextHandler{addvisit})
	http.Handle("/api/getallclients", ContextHandler{getallclients})
	http.Handle("/api/getallvisits/", ContextHandler{getallvisits})
	http.Handle("/api/users", ContextHandler{getallusers})
	http.Handle("/api/users/edit", ContextHandler{editusers})
}

var funcMap = template.FuncMap{"age": age,
	"girls":   numGirls,
	"boys":    numBoys,
	"famSize": famSize,
}
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

func numBoys(children []fammbr) int {
	n := 0
	for _, child := range children {
		if !child.Female {
			n++
		}
	}
	return n
}

func numGirls(children []fammbr) int {
	n := 0
	for _, child := range children {
		if child.Female {
			n++
		}
	}
	return n
}

func famSize(clt client) (num int, err error) {
	var men, women int = 0, 0
	if clt.Adultmales != "" {
		men, err = strconv.Atoi(clt.Adultmales)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("unable to parse %v",
				clt.Adultmales))
		}
	}
	if clt.Adultfemales != "" {
		women, err = strconv.Atoi(clt.Adultfemales)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("unable to parse %v",
				clt.Adultfemales))
		}
	}
	return men + women + len(clt.Fammbrs), nil
}

func webuserOK(c appengine.Context, w http.ResponseWriter, r *http.Request) bool {
	if !userauthenticated(c) {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return false
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return false
	}
	u := user.Current(c)
	authzed, err := userauthorized(c, u.Email)
	if err != nil {
		c.Errorf("authorization error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	if !authzed {
		c.Warningf("authorization failure: %v", u.Email)
		w.WriteHeader(http.StatusForbidden)
		return false
	}
	return true
}

func homepage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)
	l, err := user.LogoutURL(c, "http://www.svdpsm.org/")
	data := struct {
		U, LogoutUrl string
	}{
		u.Email,
		l,
	}
	err = templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func listclientspage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)

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
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)

	re, err := regexp.Compile("[0-9]+")
	if err != nil {
		http.Error(w, "unable to parse regex: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	idstr := re.FindString(r.URL.Path)
	if len(idstr) == 0 {
		c.Warningf("id missing in path")
		http.Error(w, "client id missing in path", http.StatusNotFound)
		return
	}
	clientid, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		c.Warningf("got error %v trying to parse id %v\n", err, clientid)
		http.Error(w, "error parsing client id "+idstr+
			" ("+err.Error()+")", http.StatusInternalServerError)
		return
	}
	key := datastore.NewKey(c, "SVDPClient", "", clientid, nil)
	var clt client
	err = datastore.Get(c, key, &clt)
	if err != nil {
		c.Warningf("got error %v on datastore get for key %v\n", err,
			key)
		http.Error(w, "unable to find client",
			http.StatusNotFound)
		return
	}

	q := datastore.NewQuery("SVDPClientVisit").Ancestor(key).Order("-Visitdate")
	visits := make([]visit, 0, 10)
	_, err = q.GetAll(c, &visits)

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		Clientrec    clientrec
		Visits       []visit
	}{
		u.Email,
		l,
		clientrec{clientid, clt},
		visits,
	}

	err = templates.ExecuteTemplate(w, "client.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newclientpage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)

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
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)

	re, err := regexp.Compile("[0-9]+")
	if err != nil {
		http.Error(w, "unable to parse regex: "+err.Error(),
			http.StatusNotFound)
		return
	}
	idstr := re.FindString(r.URL.Path)
	id, err := strconv.ParseInt(idstr, 10, 64)
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

func recordvisitpage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)

	re, err := regexp.Compile("[0-9]+")
	idstr := re.FindString(r.URL.Path)

	clientid, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid/missing client id ", http.StatusBadRequest)
		return
	}
	clientkey := datastore.NewKey(c, "SVDPClient", "", clientid, nil)
	clt := client{}
	err = datastore.Get(c, clientkey, &clt)
	if err == datastore.ErrNoSuchEntity {
		http.Error(w, "Unable to find client with id "+idstr,
			http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	vst := visit{}

	data := struct {
		U, LogoutUrl string
		Client       client
		Visit        visit
	}{
		u.Email,
		l,
		clt,
		vst,
	}
	err = templates.ExecuteTemplate(w, "recordvisit.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func edituserspage(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)

	admin, err := useradmin(c, u.Email)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !admin {
		http.Error(w, "must be admin to access users page",
			http.StatusForbidden)
		return
	}

	q := datastore.NewQuery("SVDPUser").Order("Email")
	users := make([]appuser, 0, 10)
	keys, err := q.GetAll(c, &users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp useredit
	resp.Aus = make([]appuser, len(keys))
	resp.Ids = make([]int64, len(keys))

	for i := range keys {
		resp.Aus[i] = users[i]
		resp.Ids[i] = keys[i].IntID()
	}
	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		Users        useredit
	}{
		u.Email,
		l,
		resp,
	}
	err = templates.ExecuteTemplate(w, "users.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
