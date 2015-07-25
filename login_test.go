package frederic

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "bytes"
    "strings"
    "fmt"
    "encoding/json"
    "errors"
    "reflect"

    "appengine"
    "appengine/user"
    "appengine/datastore"
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

func TestAddClient(t *testing.T) {
    inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
    if err != nil {
            t.Fatalf("Failed to create instance: %v", err)
    }
    defer inst.Close()

    data := strings.NewReader(`{"Firstname": "frederic", "Lastname": "ozanam"}`)
    req, err := inst.NewRequest("PUT", "/api/addclient", data)
    if err != nil {
            t.Fatalf("Failed to create req: %v", err)
    }
    req.Header = map[string][]string{
        "Content-Type": {"application/json"},
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
        t.Errorf("got body %v (%v), want %v", body, string(body), []byte("client ozanam, frederic added"))
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

func TestGetAllClients(t *testing.T) {
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
        //data, err := json.Marshal(newclients[i])
        //if err != nil {
        //    t.Fatalf("Failed to marshal: %v", err)
        //}
        //reqa, err := inst.NewRequest("PUT", "/api/addclient", 
        //    bytes.NewReader(data))
        //if err != nil {
        //    t.Fatalf("Failed to create req: %v", err)
        //}
        //reqa.Header = map[string][]string{
        //    "Content-Type": {"application/json"},
        //}

        //aetest.Login(&user.User{Email: "test@example.org"}, reqa)

        //wa := httptest.NewRecorder()
        //c := appengine.NewContext(reqa)

        //addclient(c, wa, reqa)

        //code := wa.Code
        //if code != http.StatusCreated {
        //    t.Errorf("got code on addclient %v, want %v", code, 
        //        http.StatusCreated)
        //}
    }

    req, err := inst.NewRequest("GET", "/api/getallclients", nil)
    if err != nil {
        t.Fatalf("Failed to create req: %v", err)
    }
    aetest.Login(&user.User{Email: "test@example.org"}, req)
    w := httptest.NewRecorder()

    c := appengine.NewContext(req)
    getallclients(c, w, req)

    code := w.Code
    if code != http.StatusOK {
        t.Errorf("got code %v, want %v", code, http.StatusOK)
    }
    
    body := w.Body.Bytes()
    createdclients := []client{}
    err = json.Unmarshal(body, &createdclients)
    if err != nil {
        t.Errorf("error unmarshaling response to getclients %v\n", err)
    }
    if len(createdclients) != len(newclients) {
        t.Errorf("got %v clients, want %v\n", len(createdclients),
            len(newclients))
    }
    for i := 0; i< len(newclients); i++ {
        found := false
        for j:= 0; j<len(createdclients); j++ {
            if reflect.DeepEqual(createdclients[j], newclients[i]) {
               found = true
               break
            }
        }
        if !found {
            t.Errorf("unable to find %v in %v",
                newclients[i], &createdclients)
        }
    }
}

func addclienttodb(clt client, inst aetest.Instance) (err error) {
    data, err := json.Marshal(clt)
    if err != nil {
        return err
    }
    
    req, err := inst.NewRequest("PUT", "/api/addclient", bytes.NewReader(data))
    if err != nil {
        return err
    }
    req.Header = map[string][]string{
        "Content-Type": {"application/json"},
    }

    aetest.Login(&user.User{Email: "test@example.org"}, req)

    w := httptest.NewRecorder()
    c := appengine.NewContext(req)

    addclient(c, w, req)

    code := w.Code
    if code != http.StatusCreated {
        return errors.New(fmt.Sprintf("got code on addclient %v, want %v",
             code, http.StatusCreated))
    }
    return nil
}
