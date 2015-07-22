package frederic

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "bytes"

    "appengine"
    "appengine/user"
    "appengine/aetest"
)

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

func TestAddClientNotAuthenticated(t *testing.T) {
    inst, err := aetest.NewInstance(nil)
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    req, err := inst.NewRequest("GET", "/api/addclient", nil)
    if err != nil {
        t.Fatalf("Failed to create req1: %v", err)
    }
    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    addclient(c, w, req)

    code := w.Code
    if code != http.StatusUnauthorized {
        t.Errorf("got code %v, want %v", code, http.StatusUnauthorized)
    }
}

func TestAddClient(t *testing.T) {
    inst, err := aetest.NewInstance(nil)
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    req, err := inst.NewRequest("GET", "/api/addclient", nil)
    if err != nil {
            t.Fatalf("Failed to create req1: %v", err)
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
    if !bytes.Equal(body, []byte("client added\n")) {
        t.Errorf("got body %v, want %v", body, []byte("client added\n"))
    }
}

func TestGetClientNotAuthenticated(t *testing.T) {
    inst, err := aetest.NewInstance(nil)
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    req, err := inst.NewRequest("GET", "/api/getclient", nil)
    if err != nil {
        t.Fatalf("Failed to create req1: %v", err)
    }
    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    addclient(c, w, req)

    code := w.Code
    if code != http.StatusUnauthorized {
        t.Errorf("got code %v, want %v", code, http.StatusUnauthorized)
    }
}

