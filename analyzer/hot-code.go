package analyzer

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func AnalyzeHotFiles(repoPath string) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatalf("Failed to open repo: %v", err)
	}

	ref, _ := r.Head()
	iter, _ := r.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime})

	var prev *object.Commit
	fileChangeCount := make(map[string]int)

	err = iter.ForEach(func(c *object.Commit) error {
		if prev != nil {
			patch, err := c.PatchContext(context.Background(), prev)
			if err != nil {
				return nil // skip error
			}
			for _, stat := range patch.Stats() {
				fileChangeCount[stat.Name]++
			}
		}
		prev = c
		return nil
	})
	if err != nil {
		log.Fatalf("Iterate failed: %v", err)
	}

	// 排序输出
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range fileChangeCount {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	fmt.Println("🔥 热点文件排行：")
	for i, kv := range sorted {
		if i >= 10 {
			break
		}
		fmt.Printf("%2d. %s (%d 次变更)\n", i+1, kv.Key, kv.Value)
	}
}

