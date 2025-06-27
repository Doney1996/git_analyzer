// File: analyzer/intelligence.go
package analyzer

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"regexp"

	"github.com/go-git/go-git/v5"
)

func AnalyzeCommitStyle(repoPath string) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatalf("Failed to open repo: %v", err)
	}

	ref, _ := r.Head()
	iter, _ := r.Log(&git.LogOptions{From: ref.Hash()})

	stylePattern := regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore)(\(.*\))?: .+`)
	okCount := 0
	badCount := 0

	_ = iter.ForEach(func(c *object.Commit) error {
		if stylePattern.MatchString(c.Message) {
			okCount++
		} else {
			badCount++
		}
		return nil
	})

	total := okCount + badCount
	fmt.Printf("✅ 符合提交规范的数量: %d\n", okCount)
	fmt.Printf("⚠️ 不符合提交规范的数量: %d\n", badCount)
	fmt.Printf("📊 合规率: %.2f%%\n", float64(okCount)/float64(total)*100)
}

