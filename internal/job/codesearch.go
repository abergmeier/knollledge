package job

import (
	"context"
	"crypto/sha1"
	"fmt"

	"github.com/abergmeier/knollledge/internal/github"
)

type CodeSearch struct {
	Query         string `json:"query"`
	MaxPageNumber int    `json:"max_page"`
}

func (cs *CodeSearch) Hash() string {
	hash := sha1.New()
	bs := hash.Sum([]byte(cs.Query))
	return fmt.Sprintf("%x", bs)
}

func (cs *CodeSearch) mustRun(ctx context.Context) *github.CodeSearchResult {
	c := github.NewClient(nil)
	return cs.mustRunWithClient(ctx, c)
}

func (cs CodeSearch) mustRunWithClient(ctx context.Context, c github.Client) *github.CodeSearchResult {

	codeSearch := c.CodeSearch()
	csr, _, err := codeSearch.Search(ctx, cs.Query, &github.SearchOptions{})
	if err != nil {
		panic(err)
	}

	pn := csr.PageNumber
	for pn < csr.TotalPages && pn < uint(cs.MaxPageNumber) {
		csr, _, err = codeSearch.Search(ctx, cs.Query, &github.SearchOptions{
			ListOptions: github.ListOptions{
				Page:      int(pn + 1),
				PageToken: csr.PageToken,
			},
		})
		if err != nil {
			panic(err)
		}
	}
	return csr
}
