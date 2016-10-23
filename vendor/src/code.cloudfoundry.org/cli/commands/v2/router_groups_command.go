package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
)

type RouterGroupsCommand struct {
	usage           interface{} `usage:"CF_NAME router-groups"`
	relatedCommands interface{} `related_commands:"create-domain, domains"`
}

func (_ RouterGroupsCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ RouterGroupsCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
