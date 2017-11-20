package main

import (
	"flag"
	"fmt"
	"github.com/tolleiv/webhook-glue/lib"
	"os"
	"os/signal"
	"syscall"
)

var (
	version string
	build   string
)

func main() {
	configFile := flag.String("configFile", "filters.yaml", "path to the configuration file")
	ver := flag.Bool("version", false, "prints current roxy version")
	flag.Parse()

	if *ver {
		fmt.Printf("%s - %s\n", version, build)
		os.Exit(0)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	ch := make(chan lib.Action, 10)

	e := EventBroker{}
	e.Initialize()
	go e.Run(":8081")

	b := Backend{}
	b.Initialize(*configFile, ch, e.Notifier)
	go b.Run()

	a := App{}
	a.Initialize(*configFile, ch, e.Notifier)
	go a.Run(":8080")

	<-done
	fmt.Println("exiting")
}
