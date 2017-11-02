package main

import (
	"flag"
	"github.com/tolleiv/webhook-glue/lib"
)

func main() {
	configFile := flag.String("configFile", "filters.yaml", "path to the configuration file")
	flag.Parse()

	ch := make(chan lib.Action, 10)

	b := Backend{}
	b.Initialize(*configFile, ch)
	go b.Run()

	a := App{}
	a.Initialize(*configFile, ch)
	a.Run(":8080")
}
