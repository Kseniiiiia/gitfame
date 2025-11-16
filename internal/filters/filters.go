package filters

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

//go:embed language_extensions.json
var langExtensionsData []byte

type langEntry struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

func loadLangMap() (map[string][]string, error) {
	var entries []langEntry
	if err := json.Unmarshal(langExtensionsData, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse language_extensions.json: %w", err)
	}
	m := make(map[string][]string)
	for _, e := range entries {
		m[strings.ToLower(e.Name)] = e.Extensions
	}
	return m, nil
}

func matchAny(patterns []string, name string) (bool, error) {
	for _, pat := range patterns {
		match, err := filepath.Match(pat, name)
		if err != nil {
			return false, fmt.Errorf("invalid pattern %q: %w", pat, err)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func Apply(files []string, exts, langs, exclude, restrict []string) ([]string, error) {
	langMap, err := loadLangMap()
	if err != nil {
		return nil, err
	}

	allowedExts := make(map[string]bool)
	for _, e := range exts {
		allowedExts[strings.ToLower(e)] = true
	}
	for _, lang := range langs {
		if es, ok := langMap[strings.ToLower(lang)]; ok {
			for _, e := range es {
				allowedExts[strings.ToLower(e)] = true
			}
		}
	}

	out := make([]string, 0, len(files))
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))

		if len(exts) > 0 || len(langs) > 0 {
			if !allowedExts[ext] {
				continue
			}
		}

		if excluded, err := matchAny(exclude, f); err != nil {
			return nil, fmt.Errorf("exclude pattern error: %w", err)
		} else if excluded {
			continue
		}

		if len(restrict) > 0 {
			included, err := matchAny(restrict, f)
			if err != nil {
				return nil, fmt.Errorf("restrict-to pattern error: %w", err)
			}
			if !included {
				continue
			}
		}

		out = append(out, f)
	}
	return out, nil
}
