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

func clearChannel() {
	for len(ch) > 0 {
		<-ch
	}
}

var a App
var ch chan lib.Action

func TestMain(m *testing.M) {
	ch = make(chan lib.Action, 10)
	a = App{}
	a.Initialize("test/empty.yaml", ch, nil)
	code := m.Run()
	close(ch)
	os.Exit(code)
}

func TestEmptyFilters(t *testing.T) {
	req, _ := http.NewRequest("GET", "/filters", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "null" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestReloadFilters(t *testing.T) {
	a.Initialize("test/empty.yaml", ch, nil)

	a.ConfigFile = "test/onerule.yaml"

	req, _ := http.NewRequest("POST", "/reload", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusSeeOther, response.Code)

	req, _ = http.NewRequest("GET", "/filters", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body == "null" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestMatchingFilters(t *testing.T) {
	a.Initialize("test/onerule.yaml", ch, nil)
	body, _ := os.Open("test/req-staging.json")

	clearChannel()
	req, _ := http.NewRequest("POST", "/webhook", body)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	if len(ch) != 1 {
		t.Errorf("Matching filters should produce one action, found %d", len(ch))
	}
	clearChannel()
}

func TestMismatchedFilters(t *testing.T) {
	a.Initialize("test/onerule.yaml", ch, nil)
	body, _ := os.Open("test/req-production.json")

	clearChannel()
	req, _ := http.NewRequest("POST", "/webhook", body)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	if len(ch) != 0 {
		t.Errorf("Matching filters should produce no action, found %d", len(ch))
	}
	clearChannel()
}

func BenchmarkMatchingFilters(b *testing.B) {
	a.Initialize("test/onerule.yaml", ch, nil)
	body, _ := os.Open("test/req-staging.json")

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/webhook", body)
		executeRequest(req)
		clearChannel()
	}
}

func BenchmarkMismatchedFilters(b *testing.B) {
	a.Initialize("test/onerule.yaml", ch, nil)
	body, _ := os.Open("test/req-production.json")

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/webhook", body)
		executeRequest(req)
		clearChannel()
	}
}
