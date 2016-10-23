package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type BindStagingSecurityGroupCommand struct {
	RequiredArgs    flags.SecurityGroup `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME bind-staging-security-group SECURITY_GROUP"`
	relatedCommands interface{}         `related_commands:"apps, bind-running-security-group, bind-security-group, restart, security-groups, staging-security-groups"`
}

func (_ BindStagingSecurityGroupCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ BindStagingSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
