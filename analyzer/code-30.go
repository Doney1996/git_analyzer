package analyzer

import (
	"bufio"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func AnalyzeRecentTopAuthors(repoPath string, days int64) error {
	cmd := exec.Command("git", "-C", repoPath, "log",
		fmt.Sprintf("--since=%d days ago", days),
		"--numstat", "--format=%H|%an|%ae", "--no-merges")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("无法获取 git 输出: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("无法执行 git log: %v", err)
	}

	type AuthorStat struct {
		Name      string
		Email     string
		Additions int
		Deletions int
		Commits   int
	}

	stats := make(map[string]*AuthorStat)
	scanner := bufio.NewScanner(stdout)

	var currentName, currentEmail string
	commitFiles := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "|") {
			commitFiles = make(map[string]bool)
			parts := strings.Split(line, "|")
			if len(parts) == 3 {
				currentName = parts[1]
				currentEmail = parts[2]
			}
		} else if line != "" {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				add, err1 := strconv.Atoi(fields[0])
				del, err2 := strconv.Atoi(fields[1])
				if err1 == nil && err2 == nil {
					key := currentEmail
					if stats[key] == nil {
						stats[key] = &AuthorStat{Name: currentName, Email: currentEmail}
					}
					stats[key].Additions += add
					stats[key].Deletions += del

					// 仅统计一次 commit 次数
					if !commitFiles["__counted__"] {
						stats[key].Commits++
						commitFiles["__counted__"] = true
					}
				}
			}
		}
	}
	_ = cmd.Wait()

	// 输出
	fmt.Printf("🕒 最近 %d 天提交活跃统计（按代码行数排序）:\n", days)
	fmt.Printf("%-20s %-6s %-6s %-6s\n", "Author", "Add", "Del", "Commits")
	fmt.Println("--------------------------------------------------")

	// 排序输出
	type pair struct {
		Key  string
		Stat *AuthorStat
	}
	var all []pair
	for k, v := range stats {
		all = append(all, pair{k, v})
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Stat.Additions > all[j].Stat.Additions
	})

	for _, p := range all {
		fmt.Printf("%-20s %-6d %-6d %-6d\n", p.Stat.Email, p.Stat.Additions, p.Stat.Deletions, p.Stat.Commits)
	}

	return nil
}

