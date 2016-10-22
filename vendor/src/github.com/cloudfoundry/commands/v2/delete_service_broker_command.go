package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type DeleteServiceBrokerCommand struct {
	RequiredArgs    flags.ServiceBroker `positional-args:"yes"`
	Force           bool                `short:"f" description:"Force deletion without confirmation"`
	usage           interface{}         `usage:"CF_NAME delete-service-broker SERVICE_BROKER [-f]"`
	relatedCommands interface{}         `related_commands:"delete-service, purge-service-offering, service-brokers"`
}

func (_ DeleteServiceBrokerCommand) Setup(config commands.Config, ui commands.UI) error {
	return nil
}

func (_ DeleteServiceBrokerCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}
