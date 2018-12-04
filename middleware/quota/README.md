# API Rate / Quota Limiter Middleware

This package provides a simple and minimalist implementation of API rate / quota request limiter within a http middleware.

## How To Use

1. Implement `APIQuotaInterface` in your package

    ```go
    type APIQuotaInterface interface {
        // Parse token from http request.
        ParseTokenFromRequest(ctx context.Context, r *http.Request) (string, error)
        // Check current token's quota plan.
        GetTokenQuotaPlan(ctx context.Context, token string) (int64, error)
    }
    ```

2. Initiate package (as singleton)

    ```go
    type APIQuotaOption struct {
        KeyPrefix string
        Cache     *redis.Pool
        Logger    *logrus.Logger
        Interface APIQuotaInterface
    }

    // Call this once.
    quota.Init(APIQuotaOption{
        KeyPrefix: "foo",
        Cache:     cachePool,
        Logger:    logrus.New(),
        Interface: customAPIQuotaImplementation{}, // Use struct defined from 1st step.
    })
    ```

## NOTE

- Ideally `GetTokenQuotaPlan` is implemented by selecting into a database.
- Expiration / validity of the API token should be managed yourself (i.e. create custom cron to set / expire tokens).

## Example

Example can be found in the test file: [`quota_test.go`](./quota_test.go).
