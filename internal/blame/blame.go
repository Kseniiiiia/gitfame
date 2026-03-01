package blame

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"unicode"
)

func runGit(repo string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %v failed: %w", args, err)
	}
	return out, nil
}

func ListFiles(repo, rev string) ([]string, error) {
	out, err := runGit(repo, "ls-tree", "-r", "--name-only", rev)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	files := make([]string, 0, len(lines))
	for _, f := range lines {
		if f = strings.TrimSpace(f); f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

func ShowFile(repo, rev, file string) ([]byte, error) {
	return runGit(repo, "show", rev+":"+file)
}

func IsEmpty(repo, rev, file string) (bool, error) {
	content, err := ShowFile(repo, rev, file)
	if err != nil {
		return false, err
	}
	return len(content) == 0, nil
}

func LastCommit(repo, rev, file string) (string, error) {
	out, err := runGit(repo, "log", "-1", "--format=%H", rev, "--", file)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

type BlameInfo struct {
	Commit string
	Author string
}

func isHexHash(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !unicode.Is(unicode.ASCII_Hex_Digit, c) {
			return false
		}
	}
	return true
}

func extractAuthorName(line string) string {
	line = strings.TrimSpace(line)
	startEmail := strings.Index(line, "<")
	if startEmail != -1 {
		return strings.TrimSpace(line[:startEmail])
	}
	return line
}

func processBlameLine(line string, useCommitter bool, currentCommit *string, currentAuthor *string) {
	fields := strings.Fields(line)

	if len(fields) >= 3 && isHexHash(fields[0]) {
		*currentCommit = fields[0]
		*currentAuthor = ""
		return
	}

	if *currentCommit != "" && *currentAuthor == "" {
		if useCommitter && strings.HasPrefix(line, "committer ") {
			*currentAuthor = extractAuthorName(strings.TrimPrefix(line, "committer "))
		} else if !useCommitter && strings.HasPrefix(line, "author ") {
			*currentAuthor = extractAuthorName(strings.TrimPrefix(line, "author "))
		}
	}
}

func handleContentLine(line string, repo string, useCommitter bool, currentCommit *string,
	currentAuthor *string, authorCache map[string]string, blame *[]BlameInfo, lineIndex *int) error {

	if *currentAuthor == "" && *currentCommit != "" {
		if cachedAuthor, exists := authorCache[*currentCommit]; exists {
			*currentAuthor = cachedAuthor
		} else {
			author, err := AuthorOf(repo, *currentCommit, useCommitter)
			if err != nil {
				return fmt.Errorf("failed to resolve author for commit %s: %w", *currentCommit, err)
			}
			*currentAuthor = author
			authorCache[*currentCommit] = author
		}
	}

	if *currentCommit != "" && *currentAuthor != "" {
		(*blame)[*lineIndex] = BlameInfo{
			Commit: *currentCommit,
			Author: *currentAuthor,
		}
		(*lineIndex)++
	}
	return nil
}

func BlameFile(repo, rev, file string, useCommitter bool) ([]BlameInfo, error) {
	out, err := runGit(repo, "blame", "--line-porcelain", rev, "--", file)
	if err != nil {
		return nil, err
	}

	content, err := ShowFile(repo, rev, file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	blame := make([]BlameInfo, len(lines))

	lineIndex := 0
	var currentCommit, currentAuthor string
	authorCache := make(map[string]string)

	for scanner.Scan() && lineIndex < len(lines) {
		line := scanner.Text()

		if strings.HasPrefix(line, "\t") {
			if err := handleContentLine(line, repo, useCommitter, &currentCommit, &currentAuthor,
				authorCache, &blame, &lineIndex); err != nil {
				return nil, err
			}
		} else {
			processBlameLine(line, useCommitter, &currentCommit, &currentAuthor)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if lineIndex != len(lines) {
		return nil, fmt.Errorf("blame output line count (%d) does not match file line count (%d)", lineIndex, len(lines))
	}

	return blame[:lineIndex], nil
}

func AuthorOf(repo, commit string, useCommitter bool) (string, error) {
	format := "%an"
	if useCommitter {
		format = "%cn"
	}
	out, err := runGit(repo, "show", "-s", "--format="+format, commit)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
