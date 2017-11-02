package main

import (
	"github.com/tolleiv/webhook-glue/lib"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

var a App

func TestMain(m *testing.M) {
	ch := make(chan lib.Action, 10)

	a = App{}
	a.Initialize("test/empty.yaml", ch)
	code := m.Run()
	os.Exit(code)
}

func TestEmptyFilters(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "null" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}
