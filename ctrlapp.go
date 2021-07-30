package main

import (
	"github.com/edgehook/ctrlapp/cmd"
	"k8s.io/component-base/logs"
	"os"
)

func main() {
	//initial log.
	logs.InitLogs()
	defer logs.FlushLogs()

	//create app command and execute.
	command := cmd.NewAppCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
