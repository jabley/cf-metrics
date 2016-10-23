package client_test

import (
	"net/http"

	"github.com/cloudfoundry/go-ccapi/v3/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("TokenRefresher", func() {
	var (
		uaaServer      *ghttp.Server
		tokenRefresher client.TokenRefresher
	)

	BeforeEach(func() {
		uaaServer = ghttp.NewServer()
		tokenRefresher = client.NewTokenRefresher(uaaServer.URL())
	})

	AfterEach(func() {
		uaaServer.Close()
	})

	Describe("Refresh", func() {
		BeforeEach(func() {
			uaaServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token", "refresh_token=old-refresh-token&grant_type=refresh_token&scope="),
					ghttp.VerifyHeader(map[string][]string{
						"Authorization": []string{"Basic Y2Y6"},
						"accept":        []string{"application/json"},
						"content-type":  []string{"application/x-www-form-urlencoded"},
					}),
				),
			)
		})

		It("tries to refresh the auth token", func() {
			tokenRefresher.Refresh("old-refresh-token")
			Expect(uaaServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when refreshing the auth token succeeds", func() {
			BeforeEach(func() {
				uaaServer.SetHandler(0,
					ghttp.RespondWith(http.StatusCreated, `{
						"access_token": "the-access-token",
						"token_type": "the-token-type",
						"refresh_token": "the-refresh-token"
					}`),
				)
			})

			It("returns the new tokens", func() {
				accessToken, refreshToken, err := tokenRefresher.Refresh("old-refresh-token")
				Expect(err).NotTo(HaveOccurred())
				Expect(accessToken).To(Equal("the-access-token"))
				Expect(refreshToken).To(Equal("the-refresh-token"))
			})
		})

		Context("when refreshing the auth token fails because the token used to refresh is invalid", func() {
			BeforeEach(func() {
				uaaServer.SetHandler(0,
					ghttp.RespondWith(http.StatusOK, `{
						"error": "1000",
						"error_description": "error-description"
					}`),
				)
			})

			It("returns an error", func() {
				_, _, err := tokenRefresher.Refresh("old-refresh-token")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error-description"))
			})
		})

		Context("when refreshing the auth token fails because the server returns bad JSON", func() {
			BeforeEach(func() {
				uaaServer.SetHandler(0,
					ghttp.RespondWith(http.StatusOK, `:bad_json:`),
				)
			})

			It("returns an error", func() {
				_, _, err := tokenRefresher.Refresh("old-refresh-token")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
