// File: analyzer/security.go
package analyzer

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
)

var sensitivePatterns = []string{
	`(?i)apikey`, `(?i)secret`, `(?i)password`, `(?i)passwd`, `(?i)token`, `(?i)access[_-]?key`, `(?i)PRIVATE[_-]?KEY`,
}

func ScanSecurityKeywords(repoPath string) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatalf("Failed to open repo: %v", err)
	}

	ref, _ := r.Head()
	iter, _ := r.Log(&git.LogOptions{From: ref.Hash(), All: true})

	suspects := make(map[string][]string) // commit hash -> matches

	compiled := make([]*regexp.Regexp, 0)
	for _, p := range sensitivePatterns {
		compiled = append(compiled, regexp.MustCompile(p))
	}

	_ = iter.ForEach(func(c *object.Commit) error {
		for _, pattern := range compiled {
			if pattern.MatchString(c.Message) {
				suspects[c.Hash.String()] = append(suspects[c.Hash.String()], pattern.String())
			}
		}
		return nil
	})

	fmt.Println("🔐 潜在敏感信息提交：")
	for hash, matches := range suspects {
		fmt.Printf("- Commit %s 包含关键词: %s\n", hash[:8], strings.Join(matches, ", "))
	}
}

