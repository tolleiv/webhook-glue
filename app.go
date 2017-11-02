package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"io/ioutil"
	"github.com/ghodss/yaml"
	"encoding/json"
	"os"
	"github.com/gorilla/handlers"
	"github.com/tolleiv/webhook-glue/lib"
	"fmt"
)

type App struct {
	Router     *mux.Router
	Filter     []lib.Filter
	ConfigFile string
	Channel    chan<- lib.Action
}

func (a *App) Initialize(configFile string, ch chan<- lib.Action) {
	a.Channel = ch
	a.Router = mux.NewRouter()
	a.ConfigFile = configFile
	a.initializeRoutes()
	err := a.initializeFilters()
	if err != nil {
		panic(err)
	}
}

func (a *App) Run(addr string) {
	loggedRouter := handlers.LoggingHandler(os.Stdout, a.Router)
	log.Fatal(http.ListenAndServe(addr, loggedRouter))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/", a.listFilters).Methods("GET")
	a.Router.HandleFunc("/webhook", a.triggerFilters).Methods("POST")
	a.Router.HandleFunc("/reload", a.reloadFilters).Methods("POST")
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

func (a *App) listFilters(w http.ResponseWriter, r *http.Request) {
	response, _ := json.MarshalIndent(a.Filter, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (a *App) triggerFilters(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	for _, f := range a.Filter {
		if !f.Match(string(body)) {
			continue
		}
		//fmt.Printf("Matched %s\n", f.Name)

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
