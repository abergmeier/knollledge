package github

import (
	"context"
	"fmt"
	"net/http"

	qs "github.com/google/go-querystring/query"
)

type CodeSearchService service

type SearchOptions struct {
	ListOptions
}

func (s *CodeSearchService) Search(ctx context.Context, query string, opts *SearchOptions) (*CodeSearchResult, *http.Response, error) {
	result := new(CodeSearchResult)
	resp, err := s.search(ctx, &searchParameters{Query: query}, opts, result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, nil
}

func (s *CodeSearchService) search(ctx context.Context, parameters *searchParameters, opts *SearchOptions, result interface{}) (*http.Response, error) {

	params, err := qs.Values(opts)
	if err != nil {
		return nil, err
	}

	params.Set("q", parameters.Query)

	u := fmt.Sprintf("search?%s", params.Encode())

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	// req.Header.Set("Accept", strings.Join(acceptHeaders, ", "))
	return s.client.Do(ctx, req, result)
}

type CodeResult struct {
	Path      string `json:"path,omitempty"`
	Sha       string `json:"sha,omitempty"`
	RefName   string `json:"ref_name,omitempty"`
	Language  string `json:"language,omitempty"`
	RepoId    uint64 `json:"repo_id,omitempty"`
	CommitSha string `json:"commit_sha,omitempty"`
	RepoName  string `json:"repo_name,omitempty"`
}

type CodeSearchResult struct {
	Error                string        `json:"error,omitempty"`
	Failed               bool          `json:"failed,omitempty"`
	EpochId              int64         `json:"epoch_id,omitempty"`
	IndexVersion         int64         `json:"index_version,omitempty"`
	RequestId            string        `json:"request_id,omitempty"`
	ResultsCount         uint64        `json:"results_count,omitempty"`
	IsTreelightsAvail    bool          `json:"is_treelights_avail,omitempty"`
	SearchElapsedMs      uint64        `json:"search_elapsed_ms,omitempty"`
	PageToken            string        `json:"page_token,omitempty"`
	PageNumber           uint          `json:"page_number,omitempty"`
	TotalPages           uint          `json:"total_pages,omitempty"`
	ServingOffsetQueried uint64        `json:"serving_offset_queried,omitempty"`
	Results              []*CodeResult `json:"results,omitempty"`
}

type searchParameters struct {
	Query string
}
