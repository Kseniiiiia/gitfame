package stats

import "gitlab.com/slon/shad-go/gitfame/internal/output"

func compareRecords(a, b output.Record, orderBy string) bool {
	switch orderBy {
	case "commits":
		if a.Commits != b.Commits {
			return a.Commits > b.Commits
		}
		if a.Lines != b.Lines {
			return a.Lines > b.Lines
		}
		if a.Files != b.Files {
			return a.Files > b.Files
		}
	case "files":
		if a.Files != b.Files {
			return a.Files > b.Files
		}
		if a.Lines != b.Lines {
			return a.Lines > b.Lines
		}
		if a.Commits != b.Commits {
			return a.Commits > b.Commits
		}
	default:
		if a.Lines != b.Lines {
			return a.Lines > b.Lines
		}
		if a.Commits != b.Commits {
			return a.Commits > b.Commits
		}
		if a.Files != b.Files {
			return a.Files > b.Files
		}
	}
	return a.Name < b.Name
}
