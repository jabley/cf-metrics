package commands_test

import (
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/flags"

	testterm "code.cloudfoundry.org/cli/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	var (
		ui  *testterm.FakeUI
		cmd commandregistry.Command
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		deps := commandregistry.Dependency{
			UI: ui,
		}

		cmd = &commands.Version{}
		cmd.SetDependency(deps, false)
	})

	Describe("Execute", func() {
		var flagContext flags.FlagContext

		BeforeEach(func() {
			cf.Version = "5.0.0"
			cf.Name = "my-special-cf"
			cf.BuiltOnDate = "2016-02-29"
		})

		It("prints the version", func() {
			cmd.Execute(flagContext)

			Expect(ui.Outputs()).To(Equal([]string{
				"my-special-cf version 5.0.0-2016-02-29",
			}))
		})
	})
})
