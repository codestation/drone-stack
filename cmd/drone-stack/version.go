package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"strconv"
	"time"
)

var (
	// Version indicates the application version
	Version string
	// Commit indicates the git commit of the build
	Commit string
	// BuildTime indicates the date when the binary was built (set by -ldflags)
	BuildTime string
)

const versionFormatter = "drone-stack version: %s, commit: %s, built at: %s\n"

func printVersion(c *cli.Context) {
	_, _ = fmt.Fprintf(c.App.Writer, versionFormatter, Version, Commit, BuildTime)
}

func init() {
	if Version == "" {
		Version = "dev"
	}
	if Commit == "" {
		Commit = "unknown"
	}
	if BuildTime == "" {
		BuildTime = "unknown"
	} else {
		i, err := strconv.ParseInt(BuildTime, 10, 64)
		if err == nil {
			tm := time.Unix(i, 0)
			BuildTime = tm.Format("Mon Jan _2 15:04:05 2006")
		} else {
			BuildTime = "unknown"
		}
	}
}
