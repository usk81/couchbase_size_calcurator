package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// RootCmd defines root command
	RootCmd = &cobra.Command{
		Use: "cbsc",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
)

// Run runs command.
func Run() {
	RootCmd.Execute()
}

// Exit finishes a runnning action.
func Exit(err error, codes ...int) {
	var code int
	if len(codes) > 0 {
		code = codes[0]
	} else {
		code = 2
	}
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}
