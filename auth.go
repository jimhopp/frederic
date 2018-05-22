package frederic

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
	"os"
	"strings"
)

type appuser struct {
	Email   string
	IsAdmin bool
	//multi-conference support?
}

func userauthenticated(c context.Context) bool {
	u := user.Current(c)
	return u != nil
}

func userauthorized(c context.Context, email string) (bool, error) {
	if v := os.Getenv("BOOTSTRAP_USER"); v != "" && v == email {
		return true, nil
	}
	lc := strings.ToLower(email)
	q := datastore.NewQuery("SVDPUser").Filter("Email=", lc)
	cnt, err := q.Count(c)
	log.Debugf(c, "count for user %v = %v", lc, cnt)
	if err != nil {
		return false, err
	}
	if cnt > 0 {
		return true, nil
	}
	return false, nil
}

func useradmin(c context.Context, email string) (bool, error) {
	if v := os.Getenv("BOOTSTRAP_USER"); v != "" && v == email {
		// bootstrap user is admin by definition
		return true, nil
	}
	u := []appuser{}
	q := datastore.NewQuery("SVDPUser").Filter("Email=", email)
	_, err := q.GetAll(c, &u)
	log.Debugf(c, "user %v", u)
	if err != nil {
		return false, err
	}
	if len(u) == 0 {
		return false, fmt.Errorf("no user with email %v found",
			email)
	}
	return u[0].IsAdmin, nil

}
