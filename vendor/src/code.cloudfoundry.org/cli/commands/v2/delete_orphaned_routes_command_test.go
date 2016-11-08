package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/commands/commandsfakes"
	"code.cloudfoundry.org/cli/commands/v2"
	"code.cloudfoundry.org/cli/commands/v2/common"
	"code.cloudfoundry.org/cli/commands/v2/v2fakes"
	"code.cloudfoundry.org/cli/utils/configv3"
	"code.cloudfoundry.org/cli/utils/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("DeletedOrphanedRoutes Command", func() {
	var (
		cmd        v2.DeleteOrphanedRoutesCommand
		fakeUI     *ui.UI
		fakeActor  *v2fakes.FakeDeleteOrphanedRoutesActor
		fakeConfig *commandsfakes.FakeConfig
		input      *Buffer
		executeErr error
	)

	BeforeEach(func() {
		input = NewBuffer()
		out := NewBuffer()
		fakeUI = ui.NewTestUI(input, out, out)
		fakeActor = new(v2fakes.FakeDeleteOrphanedRoutesActor)
		fakeConfig = new(commandsfakes.FakeConfig)
		fakeConfig.ExperimentalReturns(true)

		cmd = v2.DeleteOrphanedRoutesCommand{
			UI:     fakeUI,
			Actor:  fakeActor,
			Config: fakeConfig,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("Displays the experimental warning message", func() {
		Expect(fakeUI.Out).To(Say(v2.ExperimentalWarning))
	})

	Context("when checking that the user is logged in, and org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.BinaryNameReturns("faceman")
		})

		It("returns an error if the check fails", func() {
			Expect(executeErr).To(MatchError(common.NotLoggedInError{
				BinaryName: "faceman",
			}))
		})
	})

	Context("when the user is logged in, and org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
			fakeConfig.AccessTokenReturns("some-access-token")
			fakeConfig.RefreshTokenReturns("some-refresh-token")
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
		})

		Context("when the '-f' flag is provided", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			It("does not prompt for user confirmation", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeUI.Out).ToNot(Say("Really delete orphaned routes\\?>> \\[yN\\]:"))
			})
		})

		Context("when the '-f' flag is not provided", func() {
			Context("when user is prompted for confirmation", func() {
				BeforeEach(func() {
					input.Write([]byte("\n"))
				})

				It("displays the interactive prompt", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeUI.Out).To(Say("Really delete orphaned routes\\?>> \\[yN\\]:"))
				})
			})

			Context("when the user inputs no", func() {
				BeforeEach(func() {
					input.Write([]byte("n\n"))
				})

				It("does not delete orphaned routes", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetOrphanedRoutesBySpaceCallCount()).To(Equal(0))
					Expect(fakeActor.DeleteRouteCallCount()).To(Equal(0))
				})
			})

			Context("when the user inputs yes", func() {
				var routes []v2actions.Route

				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(configv3.User{
						Name: "some-user",
					}, nil)

					input.Write([]byte("y\n"))

					routes = []v2actions.Route{
						{
							GUID:   "route-1-guid",
							Host:   "route-1",
							Domain: "bosh-lite.com",
							Path:   "/path",
						},
						{
							GUID:   "route-2-guid",
							Host:   "route-2",
							Domain: "bosh-lite.com",
						},
					}

					fakeActor.GetOrphanedRoutesBySpaceReturns(routes, nil, nil)
				})

				Context("when getting the current user returns an error", func() {
					var err error

					BeforeEach(func() {
						err = errors.New("getting current user error")
						fakeConfig.CurrentUserReturns(configv3.User{}, err)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(err))
					})
				})

				It("displays getting routes message", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeUI.Out).To(Say("Getting routes as some-user ...\n"))
				})

				It("deletes the routes and displays that they are deleted", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeActor.GetOrphanedRoutesBySpaceCallCount()).To(Equal(1))
					Expect(fakeActor.GetOrphanedRoutesBySpaceArgsForCall(0)).To(Equal("some-space-guid"))
					Expect(fakeActor.DeleteRouteCallCount()).To(Equal(2))
					Expect(fakeActor.DeleteRouteArgsForCall(0)).To(Equal(routes[0].GUID))
					Expect(fakeActor.DeleteRouteArgsForCall(1)).To(Equal(routes[1].GUID))

					Expect(fakeUI.Out).To(Say("Deleting route route-1.bosh-lite.com/path..."))
					Expect(fakeUI.Out).To(Say("Deleting route route-2.bosh-lite.com..."))
					Expect(fakeUI.Out).To(Say("OK"))
				})

				Context("when there are warnings", func() {
					BeforeEach(func() {
						fakeActor.GetOrphanedRoutesBySpaceReturns([]v2actions.Route{
							{GUID: "some-route-guid"},
						}, []string{"foo", "bar"}, nil)
						fakeActor.DeleteRouteReturns([]string{"baz"}, nil)
					})

					It("displays the warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeUI.Err).To(Say("foo"))
						Expect(fakeUI.Err).To(Say("bar"))
						Expect(fakeUI.Err).To(Say("baz"))
					})
				})

				Context("when getting the routes returns an error", func() {
					var expectedErr error

					Context("when the error is a DomainNotFoundError", func() {
						BeforeEach(func() {
							expectedErr = v2actions.DomainNotFoundError{}
							fakeActor.GetOrphanedRoutesBySpaceReturns(nil, nil, expectedErr)
						})

						It("should return the DomainNotFoundError", func() {
							Expect(executeErr).To(MatchError(expectedErr))
						})
					})

					Context("when the error is an OrphanedRoutesNotFoundError", func() {
						BeforeEach(func() {
							expectedErr = v2actions.OrphanedRoutesNotFoundError{}
							fakeActor.GetOrphanedRoutesBySpaceReturns(nil, nil, expectedErr)
						})

						It("should not return an error and only display 'OK'", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(fakeActor.DeleteRouteCallCount()).To(Equal(0))
						})
					})

					Context("when there is a generic error", func() {
						BeforeEach(func() {
							expectedErr = errors.New("getting orphaned routes error")
							fakeActor.GetOrphanedRoutesBySpaceReturns(nil, nil, expectedErr)
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError(expectedErr))
						})
					})
				})

				Context("when deleting a route returns an error", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("deleting route error")
						fakeActor.GetOrphanedRoutesBySpaceReturns([]v2actions.Route{
							{GUID: "some-route-guid"},
						}, nil, nil)
						fakeActor.DeleteRouteReturns(nil, expectedErr)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))
					})
				})
			})
		})
	})
})
