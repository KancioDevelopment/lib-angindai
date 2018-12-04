package quota_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AuthScureDevelopment/lib-arjuna/middleware/quota"
)

type (
	customAPIQuota struct{}

	TestVar struct {
		Scenario    string
		Token       string
		GetValue    int64
		GetCmdCount int
		IncValue    int64
		IncCmdCount int
		StatusCode  int
	}
)

var (
	mockRedisConn *redigomock.Conn = redigomock.NewConn()

	cachePool *redis.Pool = redis.NewPool(func() (redis.Conn, error) {
		return mockRedisConn, nil
	}, 10)

	tokenPlansDB = map[string]int64{
		"foo": 1000,
		"bar": 10,
	}
)

// Sample interface implementation to parse token from request. Returns error if header is empty.
func (a customAPIQuota) ParseTokenFromRequest(ctx context.Context, r *http.Request) (string, error) {
	token := r.Header.Get("Token")

	if token == "" {
		return "", errors.New("invalid empty token")
	}

	return token, nil
}

// Sample interface implementation to get token quota plan using in memory database.
// Returns error on undefined key.
func (a customAPIQuota) GetTokenQuotaPlan(ctx context.Context, token string) (int64, error) {
	plan, ok := tokenPlansDB[token]
	if !ok {
		return 0, errors.New("error no token")
	}

	return plan, nil
}

func TestHTTPMiddleware(t *testing.T) {
	quota.Init(quota.APIQuotaOption{
		KeyPrefix: "foo",
		Cache:     cachePool,
		Logger:    logrus.New(),
		Interface: customAPIQuota{},
	})

	ts := httptest.NewServer(
		quota.HTTPMiddleware(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OK"))
				},
			),
		),
	)
	defer ts.Close()

	client := http.Client{}

	var testVars = []TestVar{
		TestVar{
			Scenario:    "Test normal",
			Token:       "foo",
			GetValue:    900,
			GetCmdCount: 1,
			IncValue:    901,
			IncCmdCount: 1,
			StatusCode:  http.StatusOK,
		},
		TestVar{
			Scenario:    "Test quota over limit",
			Token:       "bar",
			GetValue:    10,
			GetCmdCount: 1,
			IncValue:    0,
			IncCmdCount: 0,
			StatusCode:  http.StatusTooManyRequests,
		},
		TestVar{
			Scenario:    "Test invalid token",
			Token:       "baz",
			GetValue:    -1,
			GetCmdCount: 1,
			IncValue:    0,
			IncCmdCount: 0,
			StatusCode:  http.StatusUnauthorized,
		},
		TestVar{
			Scenario:    "Test get token quota plan error",
			Token:       "ban",
			GetValue:    1,
			GetCmdCount: 1,
			IncValue:    0,
			IncCmdCount: 0,
			StatusCode:  http.StatusInternalServerError,
		},
		TestVar{
			Scenario:    "Test missing token in request",
			Token:       "",
			GetValue:    0,
			GetCmdCount: 0,
			IncValue:    0,
			IncCmdCount: 0,
			StatusCode:  http.StatusBadRequest,
		},
	}

	for _, v := range testVars {
		t.Log("Scenario:", v.Scenario)

		// Request preparations.
		cmdGetKey := mockRedisConn.Command("GET", quota.CacheKey(v.Token))
		if v.GetValue != -1 {
			cmdGetKey = cmdGetKey.Expect(v.GetValue)
		}

		cmdIncKey := mockRedisConn.Command("INC", quota.CacheKey(v.Token))
		if v.IncValue != -1 {
			cmdIncKey = cmdIncKey.Expect(v.IncValue)
		}

		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		require.NoError(t, err)

		req.Header.Set("Token", v.Token)

		// Do request.
		resp, err := client.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		// Read response.
		bResp, err := httputil.DumpResponse(resp, true)
		require.NoError(t, err)

		t.Log("Response:", string(bResp))

		assert.Equal(t, v.GetCmdCount, mockRedisConn.Stats(cmdGetKey), "invalid command GET executed count")
		assert.Equal(t, v.IncCmdCount, mockRedisConn.Stats(cmdIncKey), "invalid command INC executed count")
		assert.Equal(t, v.StatusCode, resp.StatusCode, "invalid response status code")

		<-time.After(time.Millisecond * 100)
	}
}
