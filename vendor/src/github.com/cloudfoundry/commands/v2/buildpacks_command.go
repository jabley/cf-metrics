package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type BuildpacksCommand struct {
	usage           interface{} `usage:"CF_NAME buildpacks"`
	relatedCommands interface{} `related_commands:"push"`
}

func (_ BuildpacksCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ BuildpacksCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
