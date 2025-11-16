package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

type Record struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

func Write(w io.Writer, records []Record, format string) error {
	switch format {
	case "tabular":
		return writeTabular(w, records)
	case "csv":
		return writeCSV(w, records)
	case "json":
		return json.NewEncoder(w).Encode(records)
	case "json-lines":
		for _, r := range records {
			line, err := json.Marshal(r)
			if err != nil {
				return err
			}
			if _, err := w.Write(append(line, '\n')); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func writeTabular(w io.Writer, r []Record) error {
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', tabwriter.TabIndent)

	if _, err := fmt.Fprintln(tw, "Name\tLines\tCommits\tFiles"); err != nil {
		return err
	}
	for _, x := range r {
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n", x.Name, x.Lines, x.Commits, x.Files); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	return nil
}

func writeCSV(w io.Writer, r []Record) error {
	c := csv.NewWriter(w)

	if err := c.Write([]string{"Name", "Lines", "Commits", "Files"}); err != nil {
		return err
	}
	for _, x := range r {
		if err := c.Write([]string{x.Name, fmt.Sprint(x.Lines), fmt.Sprint(x.Commits), fmt.Sprint(x.Files)}); err != nil {
			return err
		}
	}
	c.Flush()
	if err := c.Error(); err != nil {
		return err
	}
	return nil
}
