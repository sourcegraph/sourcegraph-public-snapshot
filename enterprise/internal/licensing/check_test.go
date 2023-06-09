package licensing

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

func Test_licenseChecker(t *testing.T) {
	// Connect to local redis for testing, this is the same URL used in rcache.SetupForTest
	store = redispool.NewKeyValue("127.0.0.1:6379", &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 5 * time.Second,
	})

	siteID := "some-site-id"
	token := "test-token"

	t.Run("skips check if license is air-gapped", func(t *testing.T) {
		var featureChecked Feature
		defaultMock := MockCheckFeature
		MockCheckFeature = func(feature Feature) error {
			featureChecked = feature
			return nil
		}

		t.Cleanup(func() {
			MockCheckFeature = defaultMock
		})

		store.Del(licenseValidityStoreKey)
		store.Del(lastCalledAtStoreKey)

		doer := &mockDoer{
			status:   '1',
			response: []byte(``),
		}
		handler := licenseChecker{
			siteID: siteID,
			token:  token,
			doer:   doer,
			logger: logtest.NoOp(t),
		}

		err := handler.Handle(context.Background())
		require.NoError(t, err)

		// check feature was checked
		require.Equal(t, FeatureAllowAirGapped, featureChecked)

		// check doer NOT called
		require.False(t, doer.DoCalled)

		// check result was set to true
		valid, err := store.Get(licenseValidityStoreKey).Bool()
		require.NoError(t, err)
		require.True(t, valid)

		// check last called at was set
		lastCalledAt, err := store.Get(lastCalledAtStoreKey).String()
		require.NoError(t, err)
		require.NotEmpty(t, lastCalledAt)
	})

	tests := map[string]struct {
		response []byte
		status   int
		want     bool
		err      bool
	}{
		"returns error if unable to make a request to license server": {
			response: []byte(`{"error": "some error"}`),
			status:   http.StatusInternalServerError,
			err:      true,
		},
		"returns error if got error": {
			response: []byte(`{"error": "some error"}`),
			status:   http.StatusOK,
			err:      true,
		},
		`returns correct result for "true"`: {
			response: []byte(`{"data": {"is_valid": true}}`),
			status:   http.StatusOK,
			want:     true,
		},
		`returns correct result for "false"`: {
			response: []byte(`{"data": {"is_valid": false, "reason": "some reason"}}`),
			status:   http.StatusOK,
			want:     false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			store.Del(licenseValidityStoreKey)
			store.Del(lastCalledAtStoreKey)

			doer := &mockDoer{
				status:   test.status,
				response: test.response,
			}
			checker := licenseChecker{
				siteID: siteID,
				token:  token,
				doer:   doer,
				logger: logtest.NoOp(t),
			}

			err := checker.Handle(context.Background())
			if test.err {
				require.Error(t, err)

				// check result was NOT set
				require.True(t, store.Get(licenseValidityStoreKey).IsNil())
			} else {
				require.NoError(t, err)

				// check result was set
				got, err := store.Get(licenseValidityStoreKey).Bool()
				require.NoError(t, err)
				require.Equal(t, test.want, got)
			}

			// check last called at was set
			lastCalledAt, err := store.Get(lastCalledAtStoreKey).String()
			require.NoError(t, err)
			require.NotEmpty(t, lastCalledAt)

			// check doer with proper parameters
			require.True(t, doer.DoCalled)
			require.Equal(t, "POST", doer.Request.Method)
			require.Equal(t, "https://sourcegraph.com/.api/license/check", doer.Request.URL.String())
			require.Equal(t, "application/json", doer.Request.Header.Get("Content-Type"))
			require.Equal(t, "Bearer "+token, doer.Request.Header.Get("Authorization"))
			var body struct {
				SiteID string `json:"siteID"`
			}
			err = json.NewDecoder(doer.Request.Body).Decode(&body)
			require.NoError(t, err)
			require.Equal(t, siteID, body.SiteID)
		})
	}
}

var strPtr = func(s string) *string { return &s }

type mockDoer struct {
	DoCalled bool
	Request  *http.Request

	status   int
	response []byte
}

func (d *mockDoer) Do(req *http.Request) (*http.Response, error) {
	d.DoCalled = true
	d.Request = req

	return &http.Response{
		StatusCode: d.status,
		Body:       io.NopCloser(bytes.NewReader(d.response)),
	}, nil
}
