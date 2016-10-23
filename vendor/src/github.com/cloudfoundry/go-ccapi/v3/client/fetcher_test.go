package client_test

import (
	"errors"
	"net/http"
	"strings"

	"github.com/cloudfoundry/go-ccapi/v3/client"

	clientfakes "github.com/cloudfoundry/go-ccapi/v3/client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Fetcher", func() {
	var (
		fetcher        client.Fetcher
		tokenRefresher *clientfakes.FakeTokenRefresher
		server         *ghttp.Server
		req            *http.Request
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		tokenRefresher = &clientfakes.FakeTokenRefresher{}
		fetcher = client.NewBaseFetcher(tokenRefresher, "refresh-token")
		var err error
		req, err = http.NewRequest("GET", server.URL()+"/the-path", strings.NewReader(""))
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Authorization", "old-access-token")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Fetch", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/the-path"),
					ghttp.RespondWith(http.StatusOK, `{"key": "value"}`),
				),
			)
		})

		It("executes the given request", func() {
			_, err := fetcher.Fetch(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("returns the resource from the server", func() {
			responseBytes, err := fetcher.Fetch(req)
			Expect(err).NotTo(HaveOccurred())
			Expect(responseBytes).To(Equal([]byte(`{"key": "value"}`)))
		})

		Context("when the server returns an invalid token response", func() {
			BeforeEach(func() {
				server.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyHeader(http.Header{
							"Authorization": []string{"old-access-token"},
						}),
						ghttp.VerifyRequest("GET", "/the-path"),
						ghttp.RespondWith(http.StatusOK, `{"code":1000}`),
					),
				)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyHeader(http.Header{
							"Authorization": []string{"new-access-token"},
						}),
						ghttp.VerifyRequest("GET", "/the-path"),
						ghttp.RespondWith(http.StatusOK, `{"key":"value"}`),
					),
				)
			})

			It("tries to refresh the auth token", func() {
				tokenRefresher.RefreshReturns("", "", errors.New("some-error")) // to stop it making the second request
				fetcher.Fetch(req)
				Expect(tokenRefresher.RefreshCallCount()).To(Equal(1))
			})

			Context("when refreshing the auth token succeeds", func() {
				BeforeEach(func() {
					tokenRefresher.RefreshReturns("new-access-token", "new-refresh-token", nil)
				})

				It("uses the new token when executing request again", func() {
					fetcher.Fetch(req)
					Expect(server.ReceivedRequests()).To(HaveLen(2))
				})

				It("returns the response from retrying the request", func() {
					responseBytes, _ := fetcher.Fetch(req)
					Expect(responseBytes).To(Equal([]byte(`{"key":"value"}`)))
				})

				Context("when another request is made after the token has been refreshed", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyHeader(http.Header{
									"Authorization": []string{"new-access-token"},
								}),
								ghttp.VerifyRequest("GET", "/the-other-path"),
								ghttp.RespondWith(http.StatusOK, `{"key":"value"}`),
							),
						)
					})

					It("uses the already-refreshed access token rather than the token on the new request", func() {
						fetcher.Fetch(req) // this is required to refresh the token

						newReq, err := http.NewRequest("GET", server.URL()+"/the-other-path", strings.NewReader(""))
						Expect(err).NotTo(HaveOccurred())

						newReq.Header.Set("Authorization", "bad-token")

						responseBytes, _ := fetcher.Fetch(newReq)
						Expect(responseBytes).To(Equal([]byte(`{"key":"value"}`)))
					})
				})
			})

			Context("when refreshing the auth token fails", func() {
				BeforeEach(func() {
					tokenRefresher.RefreshReturns("", "", errors.New("refresh-token-err"))
				})

				It("returns an error", func() {
					_, err := fetcher.Fetch(req)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Failed to refresh auth token: refresh-token-err"))
				})
			})
		})

		Context("when the server returns invalid JSON", func() {
			BeforeEach(func() {
				server.SetHandler(0,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/the-path"),
						ghttp.RespondWith(http.StatusOK, `:bad_json:`),
					),
				)
			})

			It("returns an error", func() {
				_, err := fetcher.Fetch(req)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
