package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/tolleiv/webhook-glue/lib"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// App wraps the HTTP-Application parts and the endpoints for webhooks input
type App struct {
	Router      *mux.Router
	Filter      []lib.Filter
	ConfigFile  string
	Channel     chan<- lib.Action
	EventStream chan<- []byte
}

// Initialize all external dependencies for App
func (a *App) Initialize(configFile string, ch chan<- lib.Action, e chan<- []byte) {
	a.Channel = ch
	a.EventStream = e
	a.Router = mux.NewRouter()
	a.ConfigFile = configFile
	a.initializeRoutes()
	err := a.initializeFilters()
	if err != nil {
		panic(err)
	}
}

// Run starts the HTTP server
func (a *App) Run(addr string) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, a.Router)
	log.Fatal(http.ListenAndServe(addr, loggedRouter))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/filters", a.listFilters).Methods("GET")
	a.Router.HandleFunc("/version", a.showVersion).Methods("GET")
	a.Router.HandleFunc("/webhook", a.triggerFilters).Methods("POST")
	a.Router.HandleFunc("/reload", a.reloadFilters).Methods("POST")
	a.Router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

}

func (a *App) initializeFilters() error {
	dat, err := ioutil.ReadFile(a.ConfigFile)
	if err != nil {
		return err
	}
	var f = struct {
		Filters []lib.Filter `json:"filters"`
	}{}
	err = yaml.Unmarshal(dat, &f)
	if err != nil {
		return err
	}
	a.Filter = f.Filters
	return nil
}

func (a *App) showVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Version: %s\nBuild: %s", version, build)
}

func (a *App) listFilters(w http.ResponseWriter, r *http.Request) {
	response, _ := json.MarshalIndent(a.Filter, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *App) triggerFilters(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	if a.EventStream != nil {
		a.EventStream <- body
	}
	for _, f := range a.Filter {
		if !f.Match(string(body)) {
			continue
		}
		params := make([]lib.ActionParam, 0)
		for _, v := range f.Values {
			params = append(params, lib.ActionParam{Name: v.Name, Value: v.Extract(string(body))})
		}
		for _, action := range f.Actions {
			a.Channel <- lib.Action{Name: action, Params: params}
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (a *App) reloadFilters(w http.ResponseWriter, r *http.Request) {
	err := a.initializeFilters()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v", err)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
