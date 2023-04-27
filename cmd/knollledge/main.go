package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/abergmeier/knollledge/internal/job"
)

var (
	inDir  = flag.String("in-dir", "", "")
	outDir = flag.String("out-dir", "", "")
)

func main() {

	gp := filepath.Join(*inDir, "*.json")
	matches, err := filepath.Glob(gp)
	log.Printf("Globbing %s\n", gp)

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	css := make(chan *job.CodeSearch)

	wg := sync.WaitGroup{}
	wg.Add(len(matches))

	go func() {
		defer close(css)
		wg.Wait()
	}()

	for _, m := range matches {
		go func(m string) {
			defer wg.Done()

			fileToChan(m, css)
		}(m)
	}

	for cs := range css {
		h := cs.Hash()
		op := filepath.Join(*outDir, fmt.Sprintf(h, ".json"))
		mustRun(ctx, cs, op)
	}
}

func fileToChan(m string, css chan<- *job.CodeSearch) {

	f, err := os.Open(m)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	cs := []job.CodeSearch{}
	err = dec.Decode(&cs)
	if err != nil {
		panic(err)
	}

	for _, s := range cs {
		func(s job.CodeSearch) {
			css <- &s
		}(s)
	}
}

func mustRun(ctx context.Context, cs *job.CodeSearch, file string) {

	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	job.MustRun(ctx, cs, f)
}
