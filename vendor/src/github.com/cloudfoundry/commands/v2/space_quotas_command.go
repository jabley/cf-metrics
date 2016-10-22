package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type SpaceQuotasCommand struct {
	usage           interface{} `usage:"CF_NAME space-quotas"`
	relatedCommands interface{} `related_commands:"set-space-quota"`
}

func (_ SpaceQuotasCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ SpaceQuotasCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
