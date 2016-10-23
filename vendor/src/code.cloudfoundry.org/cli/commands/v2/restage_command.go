package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type RestageCommand struct {
	RequiredArgs        flags.AppName `positional-args:"yes"`
	usage               interface{}   `usage:"CF_NAME restage APP_NAME"`
	relatedCommands     interface{}   `related_commands:"restart"`
	envCFStagingTimeout interface{}   `environmentName:"CF_STAGING_TIMEOUT" environmentDescription:"Max wait time for buildpack staging, in minutes" environmentDefault:"15"`
	envCFStartupTimeout interface{}   `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
}

func (_ RestageCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ RestageCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
