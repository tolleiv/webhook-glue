package main

import (
	"fmt"
	"github.com/tolleiv/webhook-glue/lib"
	"github.com/ghodss/yaml"
	//	"os/exec"
	"strings"
	"io/ioutil"
	"os/exec"
	"github.com/mgutz/logxi/v1"
)

type Backend struct {
	ConfigFile string
	Actions    []lib.Action
	Channel    <-chan lib.Action
}

func (b *Backend) Initialize(configFile string, ch <-chan lib.Action) {
	b.Channel = ch
	b.ConfigFile = configFile
	err := b.initializeActions()
	if err != nil {
		panic(err)
	}
}
func (b *Backend) initializeActions() error {
	dat, err := ioutil.ReadFile(b.ConfigFile)
	if err != nil {
		return err
	}
	var a = struct {
		Actions []lib.Action `json:"actions"`
	}{}
	err = yaml.Unmarshal(dat, &a)
	if err != nil {
		return err
	}
	b.Actions = a.Actions
	return nil
}
func (b *Backend) Run() {
	for a := range b.Channel {
		fmt.Printf("Found backend action: %s\n", a.Name)
		cmd := []string{}
		for _, aa := range b.Actions {
			if strings.Compare(a.Name, aa.Name) != 0 {
				continue
			}
			cmd = aa.Script
		}
		for _, p := range a.Params {
			c := fmt.Sprintf("export PARAM_%s=%s", strings.ToUpper(p.Name), strings.TrimSpace(p.Value))
			cmd = append([]string{c}, cmd...)
		}

		execCommand := strings.Join(cmd, " ; ")

		fmt.Printf("Command: %s\n", execCommand)
		result := exec.Command("sh", "-c", execCommand)
		stderr, err := result.CombinedOutput()
		if err != nil {
			log.Error("Executing command returned: ", err, stderr)
		}
	}
}
