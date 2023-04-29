package config

import "github.com/abergmeier/knollledge/internal/job"

var (
	codeSearches = make(map[string]job.MakeCodeSearchFunc, len(predefinedCodeSearches))
)

func init() {
	for k, cs := range codeSearches {
		codeSearches[k] = cs
	}
}

func mustToCodeSearch(name string, queryPrefix string) job.CodeSearch {
	csf := codeSearches[name]
	return csf(queryPrefix)
}
