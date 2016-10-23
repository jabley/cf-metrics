package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type BindRunningSecurityGroupCommand struct {
	RequiredArgs    flags.SecurityGroup `positional-args:"yes"`
	usage           interface{}         `usage:"CF_NAME bind-running-security-group SECURITY_GROUP\n\nTIP: Changes will not apply to existing running applications until they are restarted."`
	relatedCommands interface{}         `related_commands:"apps, bind-security-group, bind-staging-security-group, restart, running-security-groups, security-groups"`
}

func (_ BindRunningSecurityGroupCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ BindRunningSecurityGroupCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
