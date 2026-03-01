//go:build !solution

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/Kseniiiiia/gitfame/internal/blame"
	"github.com/Kseniiiiia/gitfame/internal/filters"
	"github.com/Kseniiiiia/gitfame/internal/output"
	"github.com/Kseniiiiia/gitfame/internal/stats"
)

var (
	flagRepository   = pflag.String("repository", ".", "Path to git repository")
	flagRevision     = pflag.String("revision", "HEAD", "Git revision")
	flagOrderBy      = pflag.String("order-by", "lines", "Sort by: lines, commits, files")
	flagUseCommitter = pflag.Bool("use-committer", false, "Use committer instead of author")
	flagFormat       = pflag.String("format", "tabular", "Output format: tabular, csv, json, json-lines")
	flagExtensions   = pflag.StringSlice("extensions", nil, "File extensions (e.g. .go,.md)")
	flagLanguages    = pflag.StringSlice("languages", nil, "Languages (e.g. go,markdown)")
	flagExclude      = pflag.StringSlice("exclude", nil, "Glob patterns to exclude")
	flagRestrictTo   = pflag.StringSlice("restrict-to", nil, "Glob patterns to restrict to")
)

func main() {
	pflag.Parse()

	if err := validateFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	absRepo, err := filepath.Abs(*flagRepository)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid repository path\n")
		os.Exit(1)
	}

	files, err := blame.ListFiles(absRepo, *flagRevision)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list git files: %v\n", err)
		os.Exit(1)
	}

	filtered, err := filters.Apply(files, *flagExtensions, *flagLanguages, *flagExclude, *flagRestrictTo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid glob pattern: %v\n", err)
		os.Exit(1)
	}

	statsMap, err := stats.Collect(absRepo, *flagRevision, filtered, *flagUseCommitter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to collect stats: %v\n", err)
		os.Exit(1)
	}

	records := stats.ToSortedRecords(statsMap, *flagOrderBy)

	if err := output.Write(os.Stdout, records, *flagFormat); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func validateFlags() error {
	switch *flagOrderBy {
	case "lines", "commits", "files":
	default:
		return fmt.Errorf("invalid --order-by value: %s", *flagOrderBy)
	}

	switch *flagFormat {
	case "tabular", "csv", "json", "json-lines":
	default:
		return fmt.Errorf("invalid --format value: %s", *flagFormat)
	}

	return nil
}
