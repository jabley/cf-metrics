package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

// WorkAroundPrefix is the flag in hole emoji
const WorkAroundPrefix = "\U000026f3"

type SetEnvCommand struct {
	RequiredArgs    flags.SetEnvironmentArgs `positional-args:"yes"`
	usage           interface{}              `usage:"CF_NAME set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE"`
	relatedCommands interface{}              `related_commands:"apps, env, restart, set-staging-environment-variable-group, set-running-environment-variable-group"`
}

func (_ SetEnvCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ SetEnvCommand) Execute(args []string) error {
	//TODO: Be sure to sanitize the WorkAroundPrefix
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
