package github

import (
	"context"
	"net/http"
)

func withContext(ctx context.Context, req *http.Request) *http.Request {
	return req.WithContext(ctx)
}
