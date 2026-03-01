package stats

import (
	"sort"

	"github.com/Kseniiiiia/gitfame/internal/blame"
	"github.com/Kseniiiiia/gitfame/internal/output"
)

type AuthorStat struct {
	Name    string
	Lines   int
	Commits map[string]bool
	Files   map[string]bool
}

func Collect(repo, rev string, files []string, useCommitter bool) (map[string]*AuthorStat, error) {
	stats := make(map[string]*AuthorStat)

	for _, f := range files {
		empty, err := blame.IsEmpty(repo, rev, f)
		if err != nil {
			return nil, err
		}

		if empty {
			commit, err := blame.LastCommit(repo, rev, f)
			if err != nil {
				return nil, err
			}
			name, err := blame.AuthorOf(repo, commit, useCommitter)
			if err != nil {
				return nil, err
			}
			s := getOrCreate(stats, name)
			s.Commits[commit] = true
			s.Files[f] = true
			continue
		}

		blameInfo, err := blame.BlameFile(repo, rev, f, useCommitter)
		if err != nil {
			return nil, err
		}

		for _, info := range blameInfo {
			s := getOrCreate(stats, info.Author)
			s.Lines++
			s.Commits[info.Commit] = true
			s.Files[f] = true
		}
	}

	return stats, nil
}

func getOrCreate(m map[string]*AuthorStat, name string) *AuthorStat {
	if _, ok := m[name]; !ok {
		m[name] = &AuthorStat{
			Name:    name,
			Commits: make(map[string]bool),
			Files:   make(map[string]bool),
		}
	}
	return m[name]
}

func ToSortedRecords(m map[string]*AuthorStat, orderBy string) []output.Record {
	records := make([]output.Record, 0, len(m))
	for _, s := range m {
		records = append(records, output.Record{
			Name:    s.Name,
			Lines:   s.Lines,
			Commits: len(s.Commits),
			Files:   len(s.Files),
		})
	}

	sort.Slice(records, func(i, j int) bool {
		return compareRecords(records[i], records[j], orderBy)
	})

	return records
}
