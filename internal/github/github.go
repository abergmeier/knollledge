package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

const (
	defaultCodeSearchURL = "https://cs.github.com/api/"
	headerRateLimit      = "X-RateLimit-Limit"
	headerRateRemaining  = "X-RateLimit-Remaining"
	headerRateReset      = "X-RateLimit-Reset"

	headerOTP = "X-GitHub-OTP"
)

type service struct {
	client Client
}

var errNonNilContext = errors.New("context must be non-nil")

type client struct {
	client *http.Client

	common        service // Reuse a single struct instead of allocating one for each service on the heap.
	CodeSearchURL *url.URL

	rateMu                  sync.Mutex
	secondaryRateLimitReset time.Time

	UserAgent string
}

type Client interface {
	CodeSearch() *CodeSearchService
	Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error)
	NewRequest(method, urlStr string, body interface{}, opts ...RequestOption) (*http.Request, error)
}

func NewClient(httpClient *http.Client) *client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	codeSearchURL, _ := url.Parse(defaultCodeSearchURL)

	c := &client{client: httpClient, CodeSearchURL: codeSearchURL}
	c.common.client = c
	return c
}

// NewTokenClient returns a new GitHub API client authenticated with the provided token.
func NewTokenClient(ctx context.Context, token string) *client {
	return NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))
}

type RequestOption func(req *http.Request)

// ListOptions specifies the optional parameters to various List methods that
// support offset pagination.
type ListOptions struct {
	// For paginated result sets, page of results to retrieve.
	Page      int    `url:"p,omitempty"`
	PageToken string `url:"pageToken,omitempty"`
}

func (c *client) CodeSearch() *CodeSearchService {
	return (*CodeSearchService)(&c.common)
}

func (c *client) NewRequest(method, urlStr string, body interface{}, opts ...RequestOption) (*http.Request, error) {
	if !strings.HasSuffix(c.CodeSearchURL.Path, "/") {
		return nil, fmt.Errorf("CodeSearchURL must have a trailing slash, but %q does not", c.CodeSearchURL)
	}

	u, err := c.CodeSearchURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	/*req.Header.Set("Accept", mediaTypeV3)
	 */
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	/*
		req.Header.Set(headerAPIVersion, defaultAPIVersion)
	*/

	for _, opt := range opts {
		opt(req)
	}

	return req, nil
}

func (c *client) BareDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	if ctx == nil {
		return nil, errNonNilContext
	}

	req = withContext(ctx, req)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		if e, ok := err.(*url.Error); ok {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = sanitizeURL(url).String()
				return nil, e
			}
		}

		return nil, err
	}

	response := resp

	err = CheckResponse(resp)
	if err != nil {
		defer resp.Body.Close()
		// Special case for AcceptedErrors. If an AcceptedError
		// has been encountered, the response's payload will be
		// added to the AcceptedError and returned.
		//
		// Issue #1022
		aerr, ok := err.(*AcceptedError)
		if ok {
			b, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return response, readErr
			}

			aerr.Raw = b
			err = aerr
		}

		// Update the secondary rate limit if we hit it.
		rerr, ok := err.(*github.AbuseRateLimitError)
		if ok && rerr.RetryAfter != nil {
			c.rateMu.Lock()
			c.secondaryRateLimitReset = time.Now().Add(*rerr.RetryAfter)
			c.rateMu.Unlock()
		}
	}
	return response, err
}

func CheckResponse(r *http.Response) error {
	if r.StatusCode == http.StatusAccepted {
		return &AcceptedError{}
	}
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &github.ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	// Re-populate error response body because GitHub error responses are often
	// undocumented and inconsistent.
	// Issue #1136, #540.
	r.Body = io.NopCloser(bytes.NewBuffer(data))
	switch {
	case r.StatusCode == http.StatusUnauthorized && strings.HasPrefix(r.Header.Get(headerOTP), "required"):
		return (*github.TwoFactorAuthError)(errorResponse)
	case r.StatusCode == http.StatusForbidden && r.Header.Get(headerRateRemaining) == "0":
		return &github.RateLimitError{
			Response: errorResponse.Response,
			Message:  errorResponse.Message,
		}
	case r.StatusCode == http.StatusForbidden &&
		(strings.HasSuffix(errorResponse.DocumentationURL, "#abuse-rate-limits") ||
			strings.HasSuffix(errorResponse.DocumentationURL, "#secondary-rate-limits")):
		abuseRateLimitError := &github.AbuseRateLimitError{
			Response: errorResponse.Response,
			Message:  errorResponse.Message,
		}
		if v := r.Header["Retry-After"]; len(v) > 0 {
			// According to GitHub support, the "Retry-After" header value will be
			// an integer which represents the number of seconds that one should
			// wait before resuming making requests.
			retryAfterSeconds, _ := strconv.ParseInt(v[0], 10, 64) // Error handling is noop.
			retryAfter := time.Duration(retryAfterSeconds) * time.Second
			abuseRateLimitError.RetryAfter = &retryAfter
		}
		return abuseRateLimitError
	default:
		return errorResponse
	}
}

func (c *client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return resp, err
}

func sanitizeURL(uri *url.URL) *url.URL {
	if uri == nil {
		return nil
	}
	params := uri.Query()
	if len(params.Get("client_secret")) > 0 {
		params.Set("client_secret", "REDACTED")
		uri.RawQuery = params.Encode()
	}
	return uri
}

type AcceptedError struct {
	// Raw contains the response body.
	Raw []byte
}

func (*AcceptedError) Error() string {
	return "job scheduled on GitHub side; try again later"
}

// Is returns whether the provided error equals this error.
func (ae *AcceptedError) Is(target error) bool {
	v, ok := target.(*AcceptedError)
	if !ok {
		return false
	}
	return bytes.Equal(ae.Raw, v.Raw)
}
