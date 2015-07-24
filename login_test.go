package frederic

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "bytes"

    "appengine"
    "appengine/user"
    "appengine/datastore"
    "appengine/aetest"
)

type APITest struct {
    url     string
    handler func(appengine.Context, http.ResponseWriter, *http.Request)
}

var endpoints = []APITest{{"/api/addclient", addclient}, {"/api/getallclients", getallclients}}

func TestHomePageNotLoggedIn(t *testing.T) {
    inst, err := aetest.NewInstance(nil)
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    req, err := inst.NewRequest("GET", "/", nil)
    if err != nil {
            t.Fatalf("Failed to create req1: %v", err)
    }
    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    handler(c, w, req)

    code := w.Code
    if code != http.StatusFound {
        t.Errorf("got code %v, want %v", code, http.StatusFound)
    }
}

func TestHomePageLoggedIn(t *testing.T) {
    inst, err := aetest.NewInstance(nil)
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    req, err := inst.NewRequest("GET", "/", nil)
    if err != nil {
            t.Fatalf("Failed to create req1: %v", err)
    }

    aetest.Login(&user.User{Email: "test@example.org"}, req)

    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    handler(c, w, req)

    code := w.Code
    if code != http.StatusOK {
        t.Errorf("got code %v, want %v", code, http.StatusOK)
    }

    body := w.Body.Bytes()
    if !bytes.Equal(body, []byte("This is the SVdP Clients homepage.\n\nYou are authenticated as test@example.org")) {
        t.Errorf("got body %v, want %v", body, []byte("This is the SVdP Clients homepage.\n\nYou are authenticated as test@example.org"))
    }
}

func TestEndpointsNotAuthenticated(t *testing.T) {
    inst, err := aetest.NewInstance(nil)
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    for i := 0; i < len(endpoints); i++ {
        req, err := inst.NewRequest("GET", endpoints[i].url, nil)
        if err != nil {
            t.Fatalf("Failed to create req1: %v", err)
        }
        w := httptest.NewRecorder()
        c := appengine.NewContext(req)

        endpoints[i].handler(c, w, req)

        code := w.Code
        if code != http.StatusUnauthorized {
            t.Errorf("got code %v for endpoint %v, want %v", code, 
                endpoints[i].url, http.StatusUnauthorized)
        }
    }
}

func TestAddClient(t *testing.T) {
    inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    req, err := inst.NewRequest("GET", "/api/addclient", nil)
    if err != nil {
            t.Fatalf("Failed to create req: %v", err)
    }

    aetest.Login(&user.User{Email: "test@example.org"}, req)

    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    addclient(c, w, req)

    code := w.Code
    if code != http.StatusCreated {
        t.Errorf("got code %v, want %v", code, http.StatusCreated)
    }
    body := w.Body.Bytes()
    if !bytes.Equal(body, []byte("client ozanam, frederic added")) {
        t.Errorf("got body %v, want %v", body, []byte("client ozanam, frederic added"))
    }

    q := datastore.NewQuery("SVDPClient")
    clients := make([]client, 0, 10)
    if _, err := q.GetAll(c, &clients); err != nil {
        t.Fatalf("error on GetAll: %v", err)
        return
    }
    if len(clients) != 1 {
       t.Errorf("got %v records in query, expected %v", len(clients), 1)
    }
}

func TestGetClientNotAuthenticated(t *testing.T) {
    inst, err := aetest.NewInstance(nil)
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    req, err := inst.NewRequest("GET", "/api/getallclients", nil)
    if err != nil {
        t.Fatalf("Failed to create req: %v", err)
    }
    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    addclient(c, w, req)

    code := w.Code
    if code != http.StatusUnauthorized {
        t.Errorf("got code %v, want %v", code, http.StatusUnauthorized)
    }
}

