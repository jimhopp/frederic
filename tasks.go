package main

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/user"
)

const CHUNKSIZE = 100

func startconversion(c context.Context, w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "must be admin to start conversions",
			http.StatusForbidden)
		return
	}

	qry := datastore.NewQuery("SVDPClientVisit").Limit(CHUNKSIZE)

	it := qry.Run(c)

	crsr, err := it.Cursor()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var vst visit

	_, err = it.Next(&vst)

	if err == nil {
		t := taskqueue.NewPOSTTask("/worker", url.Values{
			"cursor": {crsr.String()},
		})
		taskqueue.Add(c, t, "default")
		log.Infof(c, "added task %v", t)
	}

	l, _ := user.LogoutURL(c, "http://www.svdpsm.org/")

	data := struct {
		U, LogoutUrl string
		AddedTask    bool
	}{
		u.Email,
		l,
		err == nil,
	}

	//TODO: make this a GET of a listing of the SVDPTask records

	err = templates.ExecuteTemplate(w, "conversion.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type task struct {
	Cursor    string
	Processed int
	When      string
}

func processTask(c context.Context, w http.ResponseWriter, r *http.Request) {
	log.Infof(c, "starting task")

	crsrstr := r.FormValue("cursor")

	log.Infof(c, "cursor %v", crsrstr)

	crsr, err := datastore.DecodeCursor(crsrstr)
	if err != nil {
		log.Errorf(c, "got error %v decoding %v", err, crsrstr)
		return
	}

	it := datastore.NewQuery("SVDPClientVisit").Limit(CHUNKSIZE).Start(crsr).Run(c)

	var vst visit
	var cnt = 0

	key, err := it.Next(&vst)

	for err == nil {
		cnt++

		if !vst.Deleted {
			vst.Deleted = false
			_, err := datastore.Put(c, key, &vst)
			if err != nil {
				log.Errorf(c, "got error %v putting key %v with value %v",
					err, key, vst)
			}
		}
		key, err = it.Next(&vst)
	}

	switch {
	case err == datastore.Done && cnt > 0:
		crsr, err := it.Cursor()
		if err != nil {
			log.Errorf(c, "error getting cursor: %v", err)
			break
		}
		t := taskqueue.NewPOSTTask("/worker", url.Values{
			"cursor": {crsr.String()},
		})
		taskqueue.Add(c, t, "default")

		log.Infof(c, "added task %v", t)

		key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "SVDPTask", nil),
			&task{Cursor: crsr.String(), Processed: cnt, When: time.Now().String()})

		if err != nil {
			log.Errorf(c, "got error %v putting rec", err)
			return
		}
		log.Infof(c, "wrote rec with key %v", key)

	case err == datastore.Done:
		// we didn't have any to process

	default:
		log.Errorf(c, "got error on initial Next: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}
