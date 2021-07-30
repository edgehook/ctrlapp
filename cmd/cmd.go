package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/edgehook/ctrlapp/pkg/core"
)

/*
* new app command
 */
func NewAppCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "ctrlapp",
		Long:    `this is a advantech ctrlapp .. `,
		Version: "0.1.0",
		Run: func(cmd *cobra.Command, args []string) {
			//TODO: To help debugging, immediately log version
			klog.Infof("###########  Start the ctrlapp...! ###########")
			StartUp()
		},
	}

	return cmd
}

// StartUp.
func StartUp() {
	core.DoStartUpCore()
}
