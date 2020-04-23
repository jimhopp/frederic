package main

import (
	"os"
	"testing"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestDBEmpty(t *testing.T) {
	// t.Fatal("not implemented")
	ctx, done, err := aetest.NewContext()

	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer done()

	var kinds []string

	kinds, err = datastore.Kinds(ctx)

	if err != nil {
		t.Fatalf("Error on Kinds: %v", err)
	}

	if len(kinds) != 0 {
		t.Errorf("db has kinds\n%v", kinds)
	}

	var visits []visit
	q := datastore.NewQuery("SVDPClientVisit")
	_, err = q.GetAll(ctx, &visits)

	if err != nil {
		t.Fatalf("Error on getall: %v", err)
	}

	if len(visits) > 0 {
		t.Errorf("expected visits to be empty but has length %v with content\n%v",
			len(visits), visits)
	}
}

func TestDBStoreAndRetrieve(t *testing.T) {
	if len(os.Getenv("TRAVIS")) > 0 {
		t.Skip("works locally but not on Travis")
	}
	ctx, done, err := aetest.NewContext()

	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer done()

	_, err = datastore.Put(ctx,
		datastore.NewIncompleteKey(ctx, "SVDPClientVisit", nil),
		&visit{Deleted: false})

	if err != nil {
		t.Fatalf("Error on put: %v", err)
	}

	var retrievedVisits []visit

	q := datastore.NewQuery("SVDPClientVisit")
	_, err = q.GetAll(ctx, &retrievedVisits)

	if len(retrievedVisits) != 1 {
		t.Errorf("expected visits to to have length 1 but has length %v with content\n%v",
			len(retrievedVisits), retrievedVisits)
	}
}

type oldvisit struct {
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

func TestDBStoreWithoutPropertyCanRetrieve(t *testing.T) {
	if len(os.Getenv("TRAVIS")) > 0 {
		t.Skip("works locally but not on Travis")
	}
	ctx, done, err := aetest.NewContext()

	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer done()

	_, err = datastore.Put(ctx,
		datastore.NewIncompleteKey(ctx, "SVDPClientVisit", nil),
		&oldvisit{Vincentians: "rec 1"})

	if err != nil {
		t.Fatalf("Error on put: %v", err)
	}

	var visits []visit

	q := datastore.NewQuery("SVDPClientVisit")
	_, err = q.GetAll(ctx, &visits)

	if len(visits) != 1 {
		t.Errorf("expected visits to have length 1 but has length %v with content\n%v",
			len(visits), visits)
	} else {
		t.Logf("vists[0].Deleted=%v", visits[0].Deleted)
	}
}

func TestDBStoreWithoutPropertyCanFilter(t *testing.T) {
	ctx, done, err := aetest.NewContext()

	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer done()

	oldvisit := oldvisit{Vincentians: "rec 1"}

	_, err = datastore.Put(ctx,
		datastore.NewIncompleteKey(ctx, "SVDPClientVisit", nil),
		&oldvisit)

	if err != nil {
		t.Fatalf("Error on put: %v", err)
	}

	retrievedVisits := make([]visit, 0)

	q := datastore.NewQuery("SVDPClientVisit").Filter("Deleted =", false)
	_, err = q.GetAll(ctx, &retrievedVisits)
	if err != nil {
		t.Fatalf("Error on GetAll: %v", err)
	}

	if len(retrievedVisits) != 0 {
		t.Errorf("expected visits (deleted=false) to be empty but has length %v with content\n%v",
			len(retrievedVisits), retrievedVisits)
	}

	retrievedVisits = make([]visit, 0)

	q = datastore.NewQuery("SVDPClientVisit").Filter("Deleted =", true)
	_, err = q.GetAll(ctx, &retrievedVisits)
	if err != nil {
		t.Fatalf("Error on GetAll: %v", err)
	}

	if len(retrievedVisits) != 0 {
		t.Errorf("expected visits (deleted=true) to be empty but has length %v with content\n%v",
			len(retrievedVisits), retrievedVisits)
	}
}

func TestDBStoreWithoutPropertyCanTestProperty(t *testing.T) {
	t.Skip("cannot filter entities with no value")

	ctx, done, err := aetest.NewContext()

	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer done()

	oldvisit := oldvisit{Vincentians: "rec 1"}

	_, err = datastore.Put(ctx,
		datastore.NewIncompleteKey(ctx, "SVDPClientVisit", nil),
		&oldvisit)

	if err != nil {
		t.Fatalf("Error on put: %v", err)
	}

	retrievedVisits := make([]visit, 0)

	q := datastore.NewQuery("SVDPClientVisit")
	_, err = q.GetAll(ctx, &retrievedVisits)
	if err != nil {
		t.Fatalf("Error on GetAll: %v", err)
	}

	if len(retrievedVisits) != 1 {
		t.Errorf("expected visits to be 1 but has length %v with content\n%v",
			len(retrievedVisits), retrievedVisits)
	}

	switch {
	case retrievedVisits[0].Deleted:
		t.Logf("Deleted is true")
	case !retrievedVisits[0].Deleted:
		t.Logf("Deleted is false")
	default:
		t.Logf("Deleted has no value")
	}
}

func TestDBConversion(t *testing.T) {
	t.Skip("skipping until I figure out how to test async tasks")
	ctx, done, err := aetest.NewContext()

	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}
	defer done()

	oldvisits := []oldvisit{
		{Vincentians: "rec 1"},
		{Vincentians: "rec 2"},
	}

	for i := 0; i < len(oldvisits); i++ {
		_, err = datastore.Put(ctx,
			datastore.NewIncompleteKey(ctx, "SVDPClientVisit", nil),
			&oldvisits[i])

		if err != nil {
			t.Fatalf("Error on put: %v", err)
		}
	}

	var retrievedVisits []visit

	q := datastore.NewQuery("SVDPClientVisit").Filter("Deleted =", false)
	_, err = q.GetAll(ctx, &retrievedVisits)

	if len(retrievedVisits) != len(oldvisits) {
		t.Errorf("expected visits to have length %v but has length %v with content\n%v",
			len(oldvisits), len(retrievedVisits), retrievedVisits)
	}
}

