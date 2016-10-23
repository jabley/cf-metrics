package client_test

import (
	"errors"
	"net/http"

	"github.com/cloudfoundry/go-ccapi/v3/client"
	clientfakes "github.com/cloudfoundry/go-ccapi/v3/client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PaginatedResourceFetcher", func() {
	var (
		baseFetcher *clientfakes.FakeFetcher
		fetcher     client.Fetcher
		req         *http.Request
	)

	BeforeEach(func() {
		baseFetcher = &clientfakes.FakeFetcher{}
		fetcher = client.NewPaginatedResourceFetcher(0, baseFetcher, "refresh-token")
		var err error
		req, err = http.NewRequest("GET", "http://example.com/path", nil)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Fetch", func() {
		It("tries to fetch stuff", func() {
			fetcher.Fetch(req)
			Expect(baseFetcher.FetchCallCount()).To(Equal(1))
			Expect(baseFetcher.FetchArgsForCall(0)).To(Equal(req))
		})

		Context("when given a limit of 0", func() {
			BeforeEach(func() {
				fetcher = client.NewPaginatedResourceFetcher(0, baseFetcher, "refresh-token")
				baseFetcher.FetchStub = func(*http.Request) ([]byte, error) {
					if baseFetcher.FetchCallCount() == 1 {
						return page1JSON, nil
					}
					return page2JSON, nil
				}
			})

			It("returns all available resources", func() {
				responseBytes, err := fetcher.Fetch(req)
				Expect(err).NotTo(HaveOccurred())
				Expect(responseBytes).To(MatchJSON(`[
						{
							"guid": "resource-1-guid"
						},
						{
							"guid": "resource-2-guid"
						},
						{
							"guid": "resource-3-guid"
						},
						{
							"guid": "resource-4-guid"
						}
					]`))
			})
		})

		Context("when given a non-zero, positive limit", func() {
			BeforeEach(func() {
				fetcher = client.NewPaginatedResourceFetcher(2, baseFetcher, "refresh-token")
				baseFetcher.FetchStub = func(*http.Request) ([]byte, error) {
					if baseFetcher.FetchCallCount() == 1 {
						return page1JSON, nil
					}
					return page2JSON, nil
				}
			})

			It("fetches only as many resources as are requested", func() {
				responseBytes, err := fetcher.Fetch(req)
				Expect(err).NotTo(HaveOccurred())
				Expect(responseBytes).To(MatchJSON(`[
						{
							"guid": "resource-1-guid"
						},
						{
							"guid": "resource-2-guid"
						}
					]`))
			})
		})

		Context("when the base fetcher returns an error", func() {
			BeforeEach(func() {
				baseFetcher.FetchReturns([]byte(`{"key":"value"}`), errors.New("fetch-err"))
			})

			It("returns an error", func() {
				_, err := fetcher.Fetch(req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("fetch-err"))
			})
		})

		Context("when the base fetcher returns invalid JSON", func() {
			BeforeEach(func() {
				baseFetcher.FetchReturns([]byte(`:bad_json:`), nil)
			})

			It("returns an error", func() {
				_, err := fetcher.Fetch(req)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var paginatedResponseJSON = []byte(`{
  "pagination": {
    "total_results": 1,
    "first": {
      "href": "/v3/apps/guid-1ba69b01-e712-4edc-b7e2-a5a6837c697d/processes?page=1&per_page=50"
    },
    "last": {
      "href": "/v3/apps/guid-1ba69b01-e712-4edc-b7e2-a5a6837c697d/processes?page=1&per_page=50"
    },
    "next": null,
    "previous": null
  },
  "resources": [
    {
      "guid": "4c2d254d-009f-4f47-b789-ece38a72c6ae",
      "type": "web",
      "command": null,
      "instances": 1,
      "memory_in_mb": 1024,
      "disk_in_mb": 1024,
      "created_at": "2015-12-22T18:28:11Z",
      "updated_at": "2015-12-22T18:28:11Z",
      "links": {
        "self": {
          "href": "/v3/processes/4c2d254d-009f-4f47-b789-ece38a72c6ae"
        },
        "scale": {
          "href": "/v3/processes/4c2d254d-009f-4f47-b789-ece38a72c6ae/scale",
          "method": "PUT"
        },
        "app": {
          "href": "/v3/apps/guid-1ba69b01-e712-4edc-b7e2-a5a6837c697d"
        },
        "space": {
          "href": "/v2/spaces/54c6dc91-2992-46bd-b8fa-cb999b370325"
        }
      }
    }
  ]
}`)

var page2JSON = []byte(`{
	"pagination": {
		"total_results": 4,
		"first": {
			"href": "/the-path?page=1&per_page=2"
		},
		"last": {
			"href": "/the-path?page=2&per_page=2"
		},
		"next": null,
		"previous": "/the-path?page=1&per_page=2"
	},
	"resources": [
		{
			"guid": "resource-3-guid"
		},
		{
			"guid": "resource-4-guid"
		}
	]
}`)

var page1JSON = []byte(`{
	"pagination": {
		"total_results": 4,
		"first": {
			"href": "/the-path?page=1&per_page=2"
		},
		"last": {
			"href": "/the-path?page=2&per_page=2"
		},
		"next": "/the-path?page=2&per_page=2",
		"previous": null
	},
	"resources": [
		{
			"guid": "resource-1-guid"
		},
		{
			"guid": "resource-2-guid"
		}
	]
}`)
