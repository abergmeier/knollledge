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

type MakeCodeSearchFunc func(queryPrefix string) CodeSearch

var (
	_ MakeCodeSearchFunc = MakeBazelPackageCodeSearch
	_ MakeCodeSearchFunc = MakeBufConfigurationCodeSearch
	_ MakeCodeSearchFunc = MakeCargoConfigurationCodeSearch
	_ MakeCodeSearchFunc = MakeContainerConfigurationCodeSearch
	_ MakeCodeSearchFunc = MakeGoModuleCodeSearch
	_ MakeCodeSearchFunc = MakePoetryConfigurationCodeSearch
	_ MakeCodeSearchFunc = MakeProtobufDefinitionCodeSearch
	_ MakeCodeSearchFunc = MakeSkaffoldConfigurationCodeSearch
	_ MakeCodeSearchFunc = MakeTerraformBackendCodeSearch
	_ MakeCodeSearchFunc = MakeTerraformGKECluster
)

func MakeBazelPackageCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:**/BUILD OR path:**/BUILD.bazel")
}

func MakeBufConfigurationCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:**/buf.yaml")
}

func MakeCargoConfigurationCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:**/Cargo.toml")
}

func MakeContainerConfigurationCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:**/Containerfile OR path:**/Dockerfile FROM")
}

func MakeGoModuleCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:**/go.mod")
}

func MakePoetryConfigurationCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:**/poetry.lock")
}

func MakeProtobufDefinitionCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:*.proto")
}

func MakeSkaffoldConfigurationCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:**/skaffold.yaml")
}

func MakeTerraformBackendCodeSearch(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, "path:*.tf backend")
}

func MakeTerraformGKECluster(queryPrefix string) CodeSearch {
	return makePrefixedCodeSearch(queryPrefix, `path:*.tf "google_container_cluster"`)
}

func makePrefixedCodeSearch(queryPrefix, query string) CodeSearch {
	return CodeSearch{
		Query: fmt.Sprint(queryPrefix, query),
	}
}
