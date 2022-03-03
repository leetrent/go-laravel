package session

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alexedwards/scs/v2"
)

func TestSession_InitSession(t *testing.T) {
	c := &Session{
		CookieLifetime: "100",
		CookiePersist:  "true",
		CookieName:     "celeritas",
		CookieDomain:   "localhost",
		SessionType:    "cookie",
	}

	var sm1 *scs.SessionManager
	sm2 := c.InitSession()

	var sessKind reflect.Kind
	var sessType reflect.Type

	rv := reflect.ValueOf(sm2)

	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		//fmt.Println("FOR LOOP:", rv.Kind(), rv.Type(), rv)
		fmt.Println()
		fmt.Printf("FOR LOOP => rv.Kind(): %v, rv.Type(): %v, rv: %v", rv.Kind(), rv.Type(), rv)
		fmt.Println()
		sessKind = rv.Kind()
		sessType = rv.Type()
		rv = rv.Elem()
	}

	if !rv.IsValid() {
		t.Error("invalid type or kind; kind:", rv.Kind(), "type:", rv.Type())
	}

	if sessKind != reflect.ValueOf(sm1).Kind() {
		t.Error("wrong kind returned testing cookie session. Expected", reflect.ValueOf(sm1).Kind(), "and got", sessKind)
	}
	if sessType != reflect.ValueOf(sm1).Type() {
		t.Error("wrong type returned testing cookie session. Expected", reflect.ValueOf(sm1).Type(), "and got", sessType)
	}
}
