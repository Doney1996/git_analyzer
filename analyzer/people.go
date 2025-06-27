package analyzer

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"time"

	"github.com/go-git/go-git/v5"
)

func AnalyzePeople(repoPath string) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatalf("Failed to open repo: %v", err)
	}

	ref, _ := r.Head()
	iter, _ := r.Log(&git.LogOptions{From: ref.Hash()})

	contributorMap := map[string]int{}
	timeMap := map[string][]time.Time{}

	err = iter.ForEach(func(c *object.Commit) error {
		name := c.Author.Name
		contributorMap[name]++
		timeMap[name] = append(timeMap[name], c.Author.When)
		return nil
	})

	fmt.Println("👨‍💻 活跃开发者统计：")
	for author, count := range contributorMap {
		nightCount := 0
		for _, t := range timeMap[author] {
			if t.Hour() >= 22 || t.Hour() < 6 {
				nightCount++
			}
		}
		fmt.Printf("- %s: %d commits, %d 夜间提交\n", author, count, nightCount)
	}
}

