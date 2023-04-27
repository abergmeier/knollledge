package github

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	expected = CodeSearchResult{
		Results: []*CodeResult{
			{
				Path:      "README.md",
				Sha:       "7f41e7bd259a78d4d1fe7e743702f146f5c73846",
				RefName:   "refs/heads/master",
				Language:  "Markdown",
				RepoId:    83222441,
				CommitSha: "a07e261677c012d37d26255de6e7b128a2643946",
				RepoName:  "donnemartin/system-design-primer",
			},
			{

				Path:      "readme.md",
				Sha:       "52ed2fb6862cc47ac4decfe8121139ead907a148",
				RefName:   "refs/heads/main",
				Language:  "Markdown",
				RepoId:    21737465,
				CommitSha: "b26d26bd1ad3e80f971edd78640d5a98b2c8e875",
				RepoName:  "sindresorhus/awesome",
			},
		},
		EpochId:              298,
		IndexVersion:         49,
		RequestId:            "A186:5796:68DC:3C8A2:64419E90",
		ResultsCount:         100,
		IsTreelightsAvail:    true,
		SearchElapsedMs:      228,
		PageToken:            "7f41e7bd52ed2fb676a8801ec3ec1ea4da297b2ea28bb37410b7da990fcee96c88d7927f3ccae2544a1deb36c08362d39ca3b87e1d529835fe6d0526238a7ae354edea4b1a2628ee11010c1bb5d384725f008f126b101fbc3db4cb06220be36411339110bbe95077fc6b61288429ee59b32511499d4903de086c73476fb794457f58d0382d4057337bb1e3e10b9d5afd763ef3b2a897c50d8b7bf729e63971ef1f6b6d7e353bbf8142cb8cffbb19f047fd894e4eaac2405fdc6295f933fce86ba5e9d938eb1f59738db6be7d329a88572aeafaeef3b843f93160dff6db70e6d6206f8ae09fce8b915fc04c87bf3786e47214f075ca2af1abab067d515cd28abdad852e4739de1d6a6f3aa94ceed0f3cb39b2a0b8eae539d6d33ab74baa59836b786c9a5471205e3d8327db0cfe26380442154e7c4ceba13ae9e3ddaa931e3c38b361d2f2cb41a2cc7dedf1109aa683e271c533596ee9884cf97f6655aa471f89f2d0b08187e911eaa3d2a37079933efbd8db1f72b906ff70255cec57fd43e0e3498a317815de279cb608ec326155a9bb",
		PageNumber:           1,
		TotalPages:           5,
		ServingOffsetQueried: 147074869,
	}
)

func TestCodeSearch(t *testing.T) {
	client := NewTestClient()
	result, _, err := client.CodeSearch().Search(context.TODO(), "path:**.java class", &SearchOptions{})
	if err != nil {
		t.Fatal("Initial search failed:", err)
	}

	result, _, err = client.CodeSearch().Search(context.TODO(), "path:**.java class", &SearchOptions{
		ListOptions: ListOptions{
			Page:      2,
			PageToken: result.PageToken,
		},
	})
	if err != nil {
		t.Fatal("Secondary search failed:", err)
	}
	if result.PageNumber != 2 {
		t.Fatal("Invalid PageNumber returned (2 expected):", result.PageNumber)
	}
}

func TestCodeSearch2(t *testing.T) {

}

func TestCodeSearchResponse(t *testing.T) {
	f, err := os.Open("testdata/response.json")
	if err != nil {
		panic(err)
	}
	dec := json.NewDecoder(f)
	result := CodeSearchResult{}
	err = dec.Decode(&result)
	if err != nil {
		t.Fatal("Decode failed:", err)
	}

	diff := cmp.Diff(result, expected)
	if diff != "" {
		t.Errorf("Decode produced unexpected result\n%s\n", diff)
	}
}
