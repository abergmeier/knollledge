package github

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

func NewTestClient() Client {
	cookiePath := filepath.Join(xdg.ConfigHome, "knollledge/cookie.combined.txt")
	f, err := os.ReadFile(cookiePath)
	if err != nil {
		panic(err)
	}

	sep := strings.ReplaceAll(string(f), "; ", "\n")

	lines := strings.Split(sep, "\n")

	c := NewClient(nil)

	tc := &testClient{
		addCookie: func(req *http.Request) {
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				vs := strings.SplitN(line, "=", 2)

				req.AddCookie(&http.Cookie{
					Name:  vs[0],
					Value: vs[1],
				})
			}
		},
		client: c,
	}
	tc.testCommon.client = tc
	return tc
}

type testClient struct {
	addCookie  RequestOption
	client     *client
	testCommon service
}

func (c *testClient) CodeSearch() *CodeSearchService {
	return (*CodeSearchService)(&c.testCommon)
}

func (c *testClient) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	return c.client.Do(ctx, req, v)
}

func (c *testClient) NewRequest(method, urlStr string, body interface{}, opts ...RequestOption) (*http.Request, error) {

	opts = append(opts, c.addCookie)
	return c.client.NewRequest(method, urlStr, body, opts...)
}
