package job

import (
	"context"
	"encoding/json"
	"io"
)

type Job interface {
	Run(ctx context.Context)
}

func MustRun(ctx context.Context, css *CodeSearch, w io.Writer) {
	res := css.mustRun(ctx)
	enc := json.NewEncoder(w)
	err := enc.Encode(res)
	if err != nil {
		panic(err)
	}
}
