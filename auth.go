package frederic

import (
	"os"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

type appuser struct {
	Email string
	//multi-conference support?
}

func userauthenticated(c appengine.Context) bool {
	u := user.Current(c)
	return u != nil
}

func userauthorized(c appengine.Context, email string) (bool, error) {
	if v := os.Getenv("BOOTSTRAP_USER"); v != "" && v == email {
		return true, nil
	}
	q := datastore.NewQuery("SVDPUser").Filter("Email=", email)
	cnt, err := q.Count(c)
	c.Debugf("count for user %v = %v", email, cnt)
	if err != nil {
		return false, err
	}
	if cnt > 0 {
		return true, nil
	}
	return false, nil
}
