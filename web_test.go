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

type EndpointTest struct {
    url     string
    handler func(appengine.Context, http.ResponseWriter, *http.Request)
    expected    int
}

var endpoints = []EndpointTest{
    {"/api/addclient", addclient, http.StatusUnauthorized}, 
    {"/api/getallclients", getallclients, http.StatusUnauthorized},
    {"/", handler, http.StatusFound},
    {"/listclients", listclients, http.StatusFound},
    {"/newclient", newclient, http.StatusFound},
}

func TestHomePage(t *testing.T) {
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
    if !bytes.Contains(body, []byte("Welcome to the home page of the Conference, test@example.org!")) {
        t.Errorf("got body %v, did not contain %v", body, []byte("Welcome to the home page of the Conference, test@example.org"))
    }

    if !bytes.Contains(body, []byte("Logout")) {
        t.Errorf("got body %v, did not contain %v", body, []byte("Logout"))
    }
}

func TestListClientsPage(t *testing.T) {
    inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    newclients := []client{
        {Firstname: "frederic", Lastname: "ozanam"},
        {Firstname: "John", Lastname: "Doe"},
        {Firstname: "Jane", Lastname: "Doe"},
    }
    for i := 0; i < len(newclients); i++ {
        err = addclienttodb(newclients[i], inst)
        if err != nil {
          t.Fatalf("unable to add client: %v", err)
        }
    }
    req, err := inst.NewRequest("GET", "/listclients", nil)
    if err != nil {
            t.Fatalf("Failed to create req: %v", err)
    }

    aetest.Login(&user.User{Email: "test@example.org"}, req)

    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    listclients(c, w, req)

    code := w.Code
    if code != http.StatusOK {
        t.Errorf("got code %v, want %v", code, http.StatusOK)
    }

    body := w.Body.Bytes()
    rows := []string{"<tr><td>First Name</td><td>Last Name</td></tr>",
        "<tr><td>frederic</td><td>ozanam</td></tr>",
        "<tr><td>John</td><td>Doe</td></tr>",
        "<tr><td>Jane</td><td>Doe</td></tr>",
    }
    for i := 0; i< len(rows); i++ {
        if !bytes.Contains(body, []byte(rows[i])) {
            t.Errorf("got body %v, did not contain %v", string(body), rows[i])
        }
    }
}

func TestAddClientPage(t *testing.T) {
    inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    req, err := inst.NewRequest("GET", "/addclient", nil)
    if err != nil {
            t.Fatalf("Failed to create req: %v", err)
    }

    aetest.Login(&user.User{Email: "test@example.org"}, req)

    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    newclient(c, w, req)

    code := w.Code
    if code != http.StatusOK {
        t.Errorf("got code %v, want %v", code, http.StatusOK)
    }

    //TODO: confirm response, create new req with filled-in values, submit?
    //      Or does this call for something like Selenium?
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
        if code != endpoints[i].expected {
            t.Errorf("got code %v for endpoint %v, want %v", code, 
                endpoints[i].url, endpoints[i].expected)
        }
    }
}

