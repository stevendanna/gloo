package main

import (
	"context"

	"github.com/solo-io/gloo/pkg/version"
	"github.com/solo-io/gloo/projects/sds/pkg/run"
	"github.com/solo-io/go-utils/contextutils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = contextutils.WithLogger(ctx, "sds_server")
	ctx = contextutils.WithLoggerValues(ctx, "version", version.Version)

	if err := run.Run(ctx, cancel); err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	}
}
