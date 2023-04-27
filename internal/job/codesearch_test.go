package job

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/abergmeier/knollledge/internal/github"
	"github.com/google/go-cmp/cmp"
)

func TestCodeSearchDecode(t *testing.T) {

	f, err := os.Open("testdata/codesearch.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	cs := []CodeSearch{}
	err = dec.Decode(&cs)
	if err != nil {
		panic(err)
	}

	diff := cmp.Diff(cs, []CodeSearch{
		{
			Query:         "path:**.java class",
			MaxPageNumber: 5,
		},
	})
	if diff != "" {
		t.Fatalf("Decode result diff:\n%s\n", diff)
	}
}

func TestMustRunCodeSearch(t *testing.T) {
	c := github.NewTestClient()
	cs := CodeSearch{
		Query: "path:**.java class",
	}
	cs.mustRunWithClient(context.TODO(), c)
}
