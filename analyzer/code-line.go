package analyzer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type StatKey struct {
	Project  string
	Name     string
	Email    string
	FileType string
}

type StatEntry struct {
	Additions int
	Deletions int
	Commits   int
}

func AnalyzeByAuthorAndFileType(repoPath, outputCSV string) error {
	cmd := exec.Command("git", "-C", repoPath, "log", "--numstat", "--format=%H|%an|%ae", "--no-merges")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("无法获取 git 输出: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("无法启动 git 命令: %v", err)
	}

	stats := make(map[StatKey]*StatEntry)
	scanner := bufio.NewScanner(stdout)

	project := filepath.Base(repoPath)
	var currentName, currentEmail string
	commitFiles := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "|") {
			// 新的 commit 开始
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
				filename := fields[2]

				if err1 == nil && err2 == nil {
					fileType := strings.ToLower(filepath.Ext(filename))
					if fileType == "" {
						fileType = "(no ext)"
					}

					key := StatKey{
						Project:  project,
						Name:     currentName,
						Email:    currentEmail,
						FileType: fileType,
					}

					if stats[key] == nil {
						stats[key] = &StatEntry{}
					}

					stats[key].Additions += add
					stats[key].Deletions += del

					if !commitFiles[fileType] {
						stats[key].Commits++
						commitFiles[fileType] = true
					}
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("git log 执行失败: %v", err)
	}

	return writeCSV(stats, outputCSV)
}

func writeCSV(stats map[StatKey]*StatEntry, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建 CSV 文件失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	writer.Write([]string{"Project", "Name", "Email", "FileType", "Additions", "Deletions", "Commits"})

	for key, entry := range stats {
		writer.Write([]string{
			key.Project,
			key.Name,
			key.Email,
			key.FileType,
			strconv.Itoa(entry.Additions),
			strconv.Itoa(entry.Deletions),
			strconv.Itoa(entry.Commits),
		})
	}

	return writer.Error()
}

func AnalyzeAndAppendCSV(repoPath, outputPath string) error {
	// 临时文件路径
	tmpFile := outputPath + ".tmp"

	// 先调用你已有的方法，生成临时 CSV
	err := AnalyzeByAuthorAndFileType(repoPath, tmpFile)
	if err != nil {
		return err
	}

	// 打开临时 CSV，跳过表头
	tmp, err := os.Open(tmpFile)
	if err != nil {
		return err
	}
	defer tmp.Close()

	scanner := bufio.NewScanner(tmp)

	var lines []string
	isFirst := true
	for scanner.Scan() {
		line := scanner.Text()
		if isFirst {
			isFirst = false
			continue // 跳过表头
		}
		lines = append(lines, line)
	}

	// 追加到目标文件（创建或打开）
	isNew := false
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		isNew = true
	}

	outFile, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	if isNew {
		writer.Write([]string{"Project", "Name", "Email", "FileType", "Additions", "Deletions", "Commits"})
	}

	for _, line := range lines {
		writer.Write(strings.Split(line, ","))
	}
	writer.Flush()

	// 删除临时文件
	_ = os.Remove(tmpFile)

	return nil
}

func AnalyzeMultipleRepos(dir, outputCsv string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("无法读取目录: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subdir := filepath.Join(dir, entry.Name())
		if _, err := os.Stat(filepath.Join(subdir, ".git")); err == nil {
			fmt.Printf("📁 正在分析: %s\n", subdir)
			err := AnalyzeAndAppendCSV(subdir, outputCsv)
			if err != nil {
				fmt.Printf("⚠️ 分析失败: %v\n", err)
			}
		}
	}
	fmt.Printf("✅ 汇总结果保存至: %s\n", outputCsv)
	return nil
}

