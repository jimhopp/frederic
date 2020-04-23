package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

func TestTasks(t *testing.T) {
	t.Skip("skip until I figure out task testing")
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("GET", "/tasks", nil)
	if err != nil {
		t.Fatalf("Failed to create req1: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	ctx := appengine.NewContext(req)
	addTestUser(ctx, "test@example.org", true)

	const numrecs = 150

	listtasks(ctx, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	rows := []string{"cnt=" + strconv.FormatInt(numrecs, 10),
		"chunks=" + strconv.FormatInt(numrecs/CHUNKSIZE, 10),
		"lastchunk=" + strconv.FormatInt(numrecs%(numrecs/CHUNKSIZE*CHUNKSIZE), 10),
	}
	for i := 0; i < len(rows); i++ {
		if !bytes.Contains(body, []byte(rows[i])) {
			t.Errorf("expected '%v' but did not find in body %v", rows[i], string(body))
		}
	}

}

func TestStartConversion(t *testing.T) {
	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer inst.Close()

	req, err := inst.NewRequest("Post", "/conversion", nil)
	if err != nil {
		t.Fatalf("Failed to create req1: %v", err)
	}

	aetest.Login(&user.User{Email: "test@example.org"}, req)

	w := httptest.NewRecorder()
	ctx := appengine.NewContext(req)
	addTestUser(ctx, "test@example.org", true)

	const numrecs = 150

	for i := 0; i < numrecs; i++ {
		_, err = datastore.Put(ctx,
			datastore.NewIncompleteKey(ctx, "SVDPClientVisit", nil),
			&oldvisit{Vincentians: "rec "})

		if err != nil {
			t.Fatalf("Error on put: %v", err)
		}
	}

	startconversion(ctx, w, req)

	code := w.Code
	if code != http.StatusOK {
		t.Errorf("got code %v, want %v", code, http.StatusOK)
	}

	body := w.Body.Bytes()
	if !bytes.Contains(body, []byte("Added task? true")) {
		t.Errorf("expected '%v' but did not find in body %v", "Added task? true", string(body))
	}

}
