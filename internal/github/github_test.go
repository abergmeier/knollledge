package github

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRequest(t *testing.T) {
	client := NewTestClient()
	req, err := client.NewRequest("GET", "search?q=path%3A%2A%2A.java+class", nil)
	if err != nil {
		t.Fatal("NewRequest failed:", err)
	}
	result := new(CodeSearchResult)
	_, err = client.Do(context.TODO(), req, result)
	if err != nil {
		t.Fatal("Do failed:", err)
	}
	expected := &CodeSearchResult{
		IsTreelightsAvail: true,
		PageNumber:        1,
		ResultsCount:      100,
	}
	diff := cmp.Diff(expected, result, cmp.FilterPath(func(p cmp.Path) bool {
		switch p.String() {
		case "EpochId":
			fallthrough
		case "IndexVersion":
			fallthrough
		case "PageToken":
			fallthrough
		case "RequestId":
			fallthrough
		case "Results":
			fallthrough
		case "ServingOffsetQueried":
			fallthrough
		case "SearchElapsedMs":
			fallthrough
		case "TotalPages":
			return true
		default:
			return false
		}
	}, cmp.Ignore()))
	if diff != "" {
		t.Fatalf("Unexpected result:\n%s\n", diff)
	}
}
