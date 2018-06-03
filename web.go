package frederic

//TODO: -figure out testing of update pages

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

type ContextHandler struct {
	Real func(context.Context, http.ResponseWriter, *http.Request)
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

type update struct {
	User string
	When string
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
	http.Handle("/editvisit/", ContextHandler{editvisitpage})
	http.Handle("/visits", ContextHandler{listvisitsinrangepage})
	http.Handle("/visitsbyclient", ContextHandler{listvisitsinrangebyclientpage})
	http.Handle("/dedupedvisits", ContextHandler{listdedupedvisitsinrangebyclientpage})
	http.Handle("/users", ContextHandler{edituserspage})
	http.Handle("/api/client", ContextHandler{addclient})
	http.Handle("/api/client/", ContextHandler{editclient})
	http.Handle("/api/visit/", ContextHandler{visitrouter})
	http.Handle("/api/getallclients", ContextHandler{getallclients})
	http.Handle("/api/getallvisits/", ContextHandler{getallvisits})
	http.Handle("/api/getvisitsinrange/", ContextHandler{getvisitsinrange})
	http.Handle("/api/users", ContextHandler{getallusers})
	http.Handle("/api/users/edit", ContextHandler{editusers})
}

var funcMap = template.FuncMap{"ages": ages,
	"girls":   numGirls,
	"boys":    numBoys,
	"add":     add,
	"famSize": famSize,
}
var templates = template.Must(template.New("client").Funcs(funcMap).ParseGlob("*.html"))

func age(dobs string) float64 {
	if len(dobs) == 0 {
		return 0
	}
	dob, err := time.Parse("2006-01-02", dobs)
	if err != nil {
		return -1
	}
	return time.Since(dob).Hours() / (24.0 * 365.25)
}

func ages(dobs string) string {
	return strconv.FormatFloat(age(dobs), 'f', 0, 64)
}

func numBoys(children []fammbr) int {
	return len(children) - numGirls(children)
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

func numMinors(children []fammbr) int {
	n := 0
	for _, child := range children {
		if age(child.DOB) < 18.0 {
			n++
		}
	}
	return n
}
func numSeniors(clt client) (num int, err error) {
	n := 0
	if a := age(clt.DOB); a >= 60.0 {
		n++
	} else if a == -1.0 {
		return 0, errors.New(fmt.Sprintf("unable to compute age from %v", clt.DOB))
	}

	for _, child := range clt.Fammbrs {
		if a := age(child.DOB); a >= 60.0 {
			n++
		} else if a == -1.0 {
			return 0, errors.New(fmt.Sprintf("unable to compute age from %v", child.DOB))
		}
	}
	return n, err
}

func numAdults(clt client) (num int, err error) {
	var men, women, seniors int = 0, 0, 0
	if clt.Adultmales != "" {
		men, err = strconv.Atoi(clt.Adultmales)
		if err != nil {
			return -1, errors.New(fmt.Sprintf("unable to parse %v",
				clt.Adultmales))
		}
	}
	if clt.Adultfemales != "" {
		women, err = strconv.Atoi(clt.Adultfemales)
		if err != nil {
			return -1, errors.New(fmt.Sprintf("unable to parse %v",
				clt.Adultfemales))
		}
	}
	if age(clt.DOB) >= 60.0 {
		seniors++
	}
	for _, child := range clt.Fammbrs {
		switch childage := age(child.DOB); {
		case childage >= 60.0:
			seniors++
		case childage >= 18.0:
			if child.Female {
				women++
			} else {
				men++
			}
		case childage < 0.0:
			return 0, errors.New(fmt.Sprintf("unable to parse %v, got %.1f", child.DOB, childage))
		}
	}

	n := men + women - seniors
	if n < 0 {
		n = 0
	}
	return n, nil
}

func famSize(clt client) (num int, err error) {
	adults, err := numAdults(clt)
	if err != nil {
		return 0, err
	}
	seniors, err := numSeniors(clt)
	if err != nil {
		return 0, err
	}
	minors := numMinors(clt.Fammbrs)
	return adults + seniors + minors, nil
}

func add(a, b int) int {
	return a + b
}

func webuserOK(c context.Context, w http.ResponseWriter, r *http.Request) bool {
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
		log.Errorf(c, "authorization error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	if !authzed {
		log.Warningf(c, "authorization failure: %v", u.Email)
		w.WriteHeader(http.StatusForbidden)
		err = templates.ExecuteTemplate(w, "unauthorized.html", nil)
		if err != nil {
			log.Errorf(c, "unauthorized user and got err on template: %v", err)
		}
		return false
	}
	return true
}

func homepage(c context.Context, w http.ResponseWriter, r *http.Request) {
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

type clientreclist []clientrec

func (c clientreclist) Len() int {
	return len(c)
}

func (c clientreclist) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c clientreclist) Less(i, j int) bool {
	if strings.ToLower(c[i].Clt.Lastname) != strings.ToLower(c[j].Clt.Lastname) {
		return strings.ToLower(c[i].Clt.Lastname) < strings.ToLower(c[j].Clt.Lastname)
	} else {
		return strings.ToLower(c[i].Clt.Firstname) < strings.ToLower(c[j].Clt.Firstname)
	}
}

func listclientspage(c context.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)

	q := datastore.NewQuery("SVDPClient")
	var clients []client
	keys, err := q.GetAll(c, &clients)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clientrecs := make(clientreclist, len(keys))
	for i := 0; i < len(clients); i++ {
		clientrecs[i].Clt = clients[i]
		clientrecs[i].Id = keys[i].IntID()
	}
	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	sort.Sort(clientrecs)

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

func getclientpage(c context.Context, w http.ResponseWriter, r *http.Request) {
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
		log.Warningf(c, "id missing in path")
		http.Error(w, "client id missing in path", http.StatusNotFound)
		return
	}
	clientid, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		log.Warningf(c, "got error %v trying to parse id %v\n", err, clientid)
		http.Error(w, "error parsing client id "+idstr+
			" ("+err.Error()+")", http.StatusInternalServerError)
		return
	}
	key := datastore.NewKey(c, "SVDPClient", "", clientid, nil)
	var clt client
	err = datastore.Get(c, key, &clt)
	if err != nil {
		log.Warningf(c, "got error %v on datastore get for key %v\n", err,
			key)
		http.Error(w, "unable to find client",
			http.StatusNotFound)
		return
	}

	q := datastore.NewQuery("SVDPClientVisit").Ancestor(key).Order("-Visitdate")
	var visits []visit
	visitkeys, err := q.GetAll(c, &visits)
	if err != nil {
		log.Warningf(c, "got error %v on datastore get for visits with key %v\n", err,
			key)
		http.Error(w, "unable to find visits",
			http.StatusInternalServerError)
		return
	}
	vstrecs := make([]visitrec, len(visitkeys))

	for i, v := range visits {
		vstrecs[i].Id = visitkeys[i].IntID()
		//not populating ClientId because they're all the same and we know it from client
		vstrecs[i].Visit = v
	}

	q = datastore.NewQuery("SVDPUpdate").Ancestor(key).Order("-When")
	var updates []update
	_, err = q.GetAll(c, &updates)

	if err != nil {
		log.Warningf(c, "got error %v on datastore get for updates with key %v\n", err,
			key)
		http.Error(w, "unable to find updates",
			http.StatusInternalServerError)
		return
	}

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		Clientrec    clientrec
		Visitrecs    []visitrec
		Updates      []update
	}{
		u.Email,
		l,
		clientrec{clientid, clt},
		vstrecs,
		updates,
	}

	err = templates.ExecuteTemplate(w, "client.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newclientpage(c context.Context, w http.ResponseWriter, r *http.Request) {
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

func editclientpage(c context.Context, w http.ResponseWriter, r *http.Request) {
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
		log.Warningf(c, "got error %v trying to parse id %v\n", err, id)
		http.Error(w, "unable to find client", http.StatusNotFound)
		return
	}
	key := datastore.NewKey(c, "SVDPClient", "", id, nil)
	var clt client
	err = datastore.Get(c, key, &clt)
	if err != nil {
		log.Warningf(c, "got error %v on datastore get for key %v\n", err,
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

func recordvisitpage(c context.Context, w http.ResponseWriter, r *http.Request) {
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
		Updates      []update
	}{
		u.Email,
		l,
		clt,
		vst,
		nil,
	}
	err = templates.ExecuteTemplate(w, "recordvisit.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func editvisitpage(c context.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	u := user.Current(c)

	re, err := regexp.Compile(`^/editvisit/([0-9]+)/([0-9]+)/edit$`)
	matches := re.FindSubmatch([]byte(r.URL.Path))
	if matches == nil {
		http.Error(w,
			fmt.Sprintf("id is missing in path for update request %v", r.URL.Path),
			http.StatusBadRequest)
		return
	}
	cltidstr := string(matches[1])
	vstidstr := string(matches[2])

	log.Debugf(c, "parsed id clt %v, vst %v from %v", cltidstr, vstidstr, r.URL.Path)

	if cltidstr == "" {
		log.Errorf(c, "cltid is missing for update request: path %v",
			r.URL.Path)
		http.Error(w,
			fmt.Sprintf("id is missing in path for update request %v", r.URL.Path),
			http.StatusBadRequest)
		return
	}

	if vstidstr == "" {
		log.Errorf(c, "vstid is missing for update request: path %v",
			r.URL.Path)
		http.Error(w,
			fmt.Sprintf("id is missing in path for update request %v", r.URL.Path),
			http.StatusBadRequest)
		return
	}

	cltid, err := strconv.ParseInt(cltidstr, 10, 64)
	if err != nil {
		log.Errorf(c, "unable to parse id %v as int64: %v", cltid, err.Error())
		http.Error(w,
			fmt.Sprintf("unable to parse id %v as int64: %v", cltid,
				err.Error()),
			http.StatusBadRequest)
		return
	}

	vstid, err := strconv.ParseInt(vstidstr, 10, 64)
	if err != nil {
		log.Errorf(c, "unable to parse vst id %v as int64: %v", vstid, err.Error())
		http.Error(w,
			fmt.Sprintf("unable to parse id %v as int64: %v", vstid,
				err.Error()),
			http.StatusBadRequest)
		return
	}

	clientkey := datastore.NewKey(c, "SVDPClient", "", cltid, nil)
	clt := client{}
	err = datastore.Get(c, clientkey, &clt)
	if err == datastore.ErrNoSuchEntity {
		http.Error(w, "Unable to find client with id "+cltidstr,
			http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	visitkey := datastore.NewKey(c, "SVDPClientVisit", "", vstid, clientkey)
	vst := visit{}
	err = datastore.Get(c, visitkey, &vst)
	if err == datastore.ErrNoSuchEntity {
		http.Error(w, "Unable to find visit with id "+vstidstr,
			http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var updates []update
	q := datastore.NewQuery("SVDPUpdate").Ancestor(visitkey).Order("-When")
	_, err = q.GetAll(c, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		U, LogoutUrl string
		Client       client
		Visit        visit
		Updates      []update
	}{
		u.Email,
		l,
		clt,
		vst,
		updates,
	}
	err = templates.ExecuteTemplate(w, "editvisit.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func edituserspage(c context.Context, w http.ResponseWriter, r *http.Request) {
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
	var users []appuser
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

	var updates []update
	q = datastore.NewQuery("SVDPUserUpdate").Order("-When")
	keys, err = q.GetAll(c, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof(c, "UserUpdates %v", updates)

	data := struct {
		U, LogoutUrl string
		Users        useredit
		Updates      []update
	}{
		u.Email,
		l,
		resp,
		updates,
	}
	err = templates.ExecuteTemplate(w, "users.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func listvisitsinrangepage(c context.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	start := r.FormValue("startdate")
	end := r.FormValue("enddate")
	csv := r.FormValue("csv")
	log.Infof(c, "looking for visits between %v and %v; csv=%v", start, end, csv)

	u := user.Current(c)

	q := datastore.NewQuery("SVDPClientVisit").
		Filter("Visitdate <=", end).
		Filter("Visitdate >=", start).
		Order("-Visitdate")
	var visits []visit
	keys, err := q.GetAll(c, &visits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cltmap := map[int64]string{}

	log.Infof(c, "got ids %v", keys)
	visitrecs := make([]visitrec, len(keys))
	for i, vst := range visits {
		visitrecs[i].Visit = vst
		visitrecs[i].Id = keys[i].IntID()
		cltkey := keys[i].Parent()
		visitrecs[i].ClientId = cltkey.IntID()
		var clt client
		err = datastore.Get(c, cltkey, &clt)
		if err != nil {
			log.Warningf(c, "unable to retrieve client with key %v for visit with key %v",
				cltkey.String(), keys[i].String())
		}
		cltmap[visitrecs[i].ClientId] = clt.Lastname + ", " + clt.Firstname
	}
	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		Visits       []visitrec
		Cltmap       map[int64]string
		Start        string
		End          string
	}{
		u.Email,
		l,
		visitrecs,
		cltmap,
		start,
		end,
	}
	if csv == "true" {
		w.Header().Set("Content-Type", "text/csv")
		err = txttemplates.ExecuteTemplate(w, "visits.csv", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		err = templates.ExecuteTemplate(w, "visits.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type visitlist []visit

func (v visitlist) Len() int {
	return len(v)
}

func (v visitlist) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v visitlist) Less(i, j int) bool {
	return v[i].Visitdate < v[j].Visitdate
}

type vstsbyclt struct {
	ClientId int64
	Name     string
	Visits   visitlist
}

type cltvsts []vstsbyclt

func (c cltvsts) Len() int {
	return len(c)
}

func (c cltvsts) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c cltvsts) Less(i, j int) bool {
	return strings.ToLower(c[i].Name) < strings.ToLower(c[j].Name)
}

func listvisitsinrangebyclientpage(c context.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	start := r.FormValue("startdate")
	end := r.FormValue("enddate")
	csv := r.FormValue("csv")
	log.Infof(c, "looking for visits between %v and %v; csv=%v", start, end, csv)

	u := user.Current(c)

	q := datastore.NewQuery("SVDPClientVisit").
		Filter("Visitdate <=", end).
		Filter("Visitdate >=", start)
	var visits []visit
	keys, err := q.GetAll(c, &visits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cltmap := map[int64]*vstsbyclt{}

	log.Infof(c, "got ids %v", keys)
	visitrecs := make([]visitrec, len(keys))
	for i, vst := range visits {
		visitrecs[i].Visit = vst
		visitrecs[i].Id = keys[i].IntID()
		cltkey := keys[i].Parent()
		visitrecs[i].ClientId = cltkey.IntID()
		var clt client
		err = datastore.Get(c, cltkey, &clt)
		if err != nil {
			log.Warningf(c, "unable to retrieve client with key %v for visit with key %v",
				cltkey.String(), keys[i].String())
		}

		var rec *vstsbyclt
		rec, ok := cltmap[visitrecs[i].ClientId]
		log.Debugf(c, "rec=%p/%v, ok=%v, cltmap=%v", rec, rec, ok, cltmap)
		if !ok {
			rec = new(vstsbyclt)
			rec.ClientId = visitrecs[i].ClientId
			rec.Name = clt.Lastname + ", " + clt.Firstname
			cltmap[visitrecs[i].ClientId] = rec
		}
		rec.Visits = append(rec.Visits, vst)
		log.Debugf(c, "rec=%v, cltmap=%v", rec, cltmap)
	}

	log.Debugf(c, "cltmao=%v", cltmap)
	var cv cltvsts
	for _, clt := range cltmap {
		sort.Sort(clt.Visits)
		cv = append(cv, *clt)
		log.Debugf(c, "appended to cltmap")
	}
	log.Debugf(c, "unsorted: cv=%v", cv)
	sort.Sort(cv)
	log.Debugf(c, "sorted: cv=%v", cv)

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		CV           cltvsts
		Start        string
		End          string
	}{
		u.Email,
		l,
		cv,
		start,
		end,
	}
	if csv == "true" {
		w.Header().Set("Content-Type", "text/csv")
		err = txttemplates.ExecuteTemplate(w, "visitsbyclient.csv", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		err = templates.ExecuteTemplate(w, "visitsbyclient.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type family struct {
	Id         int64
	Name       string
	Adults     int
	Seniors    int
	Minors     int
	FamilySize int
}

type families []family

func (f families) Len() int {
	return len(f)
}

func (f families) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f families) Less(i, j int) bool {
	return strings.ToLower(f[i].Name) < strings.ToLower(f[j].Name)
}

func listdedupedvisitsinrangebyclientpage(c context.Context, w http.ResponseWriter, r *http.Request) {
	if !webuserOK(c, w, r) {
		return
	}

	start := r.FormValue("startdate")
	end := r.FormValue("enddate")
	csv := r.FormValue("csv")
	log.Infof(c, "looking for visits between %v and %v; csv=%v", start, end, csv)

	u := user.Current(c)

	q := datastore.NewQuery("SVDPClientVisit").
		Filter("Visitdate <=", end).
		Filter("Visitdate >=", start).
		KeysOnly()
	keys, err := q.GetAll(c, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fammap := map[int64]*family{}

	var sumAdults, sumSeniors, sumMinors, sumFamSize int

	log.Debugf(c, "got ids %v", keys)
	for _, k := range keys {
		cltkey := k.Parent()
		cltId := cltkey.IntID()

		var rec *family
		rec, ok := fammap[cltId]
		log.Debugf(c, "rec=%p/%v, ok=%v, fammap=%v", rec, rec, ok, fammap)
		if !ok {
			rec = new(family)
			rec.Id = cltId
			var clt client = *new(client)
			err = datastore.Get(c, cltkey, &clt)
			if err != nil {
				log.Warningf(c, "unable to retrieve client with key %v for visit with key %v",
					cltkey.String(), k.String())
				continue
			}
			rec.Name = clt.Lastname + `, ` + clt.Firstname
			rec.Adults, err = numAdults(clt)
			if err != nil {
				log.Warningf(c, "error getting numAdults for clt %v: %v", clt, err)
				rec.Adults = 0
			}
			rec.Seniors, err = numSeniors(clt)
			if err != nil {
				log.Warningf(c, "error getting numSeniors for clt %v: %v", clt, err)
				rec.Seniors = 0
			}
			rec.Minors = numMinors(clt.Fammbrs)
			rec.FamilySize, err = famSize(clt)
			if err != nil {
				log.Warningf(c, "error getting famSize for clt %v: %v", clt, err)
				rec.Seniors = 0
			}

			fammap[cltId] = rec
			sumAdults += rec.Adults
			sumSeniors += rec.Seniors
			sumMinors += rec.Minors
			sumFamSize += rec.FamilySize
		}
		log.Debugf(c, "rec=%v, fammap=%v", rec, fammap)
	}

	log.Debugf(c, "fammap=%v", fammap)
	var cv families
	for _, f := range fammap {
		cv = append(cv, *f)
		log.Debugf(c, "appended to fammap")
	}
	log.Debugf(c, "unsorted: cv=%v", cv)
	sort.Sort(cv)
	log.Debugf(c, "sorted: cv=%v", cv)

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		CV           families
		TotalAdults  int
		TotalSeniors int
		TotalMinors  int
		TotalFamSize int
		Start        string
		End          string
	}{
		u.Email,
		l,
		cv,
		sumAdults,
		sumSeniors,
		sumMinors,
		sumFamSize,
		start,
		end,
	}
	if csv == "true" {
		w.Header().Set("Content-Type", "text/csv")
		err = txttemplates.ExecuteTemplate(w, "dedupedvisits.csv", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		err = templates.ExecuteTemplate(w, "dedupedvisits.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
