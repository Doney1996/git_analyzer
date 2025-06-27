package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/Doney1996/git_analyzer/analyzer"
)

func main() {
	var repo string
	var csv string
	var day int64

	rootCmd := &cobra.Command{
		Use:   "git-analyst.exe",
		Short: "Git Intelligence Analyzer CLI",
	}

	rootCmd.PersistentFlags().StringVar(&repo, "repo", "", "Path to Git repository (required)")
	rootCmd.PersistentFlags().StringVar(&csv, "csv", "", "csv for result")
	rootCmd.MarkPersistentFlagRequired("repo")
	rootCmd.PersistentFlags().Int64Var(&day, "day", 30, "How long to analyze the recent top authors")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "code-line-top",
		Short: "统计最近多少天内，代码提交最多的开发者",
		Run: func(cmd *cobra.Command, args []string) {
			err := analyzer.AnalyzeRecentTopAuthors(repo, day)
			if err != nil {
				log.Fatalf("执行失败: %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "people",
		Short: "分析开发者活跃度、夜间提交等",
		Run: func(cmd *cobra.Command, args []string) {
			analyzer.AnalyzePeople(repo)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "hot",
		Short: "分析热点文件",
		Run: func(cmd *cobra.Command, args []string) {
			analyzer.AnalyzeHotFiles(repo)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "style",
		Short: "分析提交规范一致性",
		Run: func(cmd *cobra.Command, args []string) {
			analyzer.AnalyzeCommitStyle(repo)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "secure",
		Short: "扫描潜在敏感信息提交",
		Run: func(cmd *cobra.Command, args []string) {
			analyzer.ScanSecurityKeywords(repo)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "line",
		Short: "统计代码行数",
		Run: func(cmd *cobra.Command, args []string) {
			// 校验目录是否存在
			if stat, err := os.Stat(repo); err != nil || !stat.IsDir() {
				log.Fatalf("指定的目录无效: %v", repo)
			}

			if len(csv) < 1 {
				log.Fatalf("csv 无效: %v", csv)
			}

			err := analyzer.AnalyzeByAuthorAndFileType(repo, csv)
			if err != nil {
				fmt.Errorf("execute error %v", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "multi-line",
		Short: "批量分析目录下所有 Git 仓库的代码行数",
		Run: func(cmd *cobra.Command, args []string) {
			if len(csv) < 1 {
				log.Fatalf("csv 无效: %v", csv)
			}
			// 校验目录是否存在
			if stat, err := os.Stat(repo); err != nil || !stat.IsDir() {
				log.Fatalf("指定的目录无效: %v", repo)
			}

			err := analyzer.AnalyzeMultipleRepos(repo, csv)
			if err != nil {
				log.Fatalf("分析失败: %v", err)
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		log.Println("执行失败:", err)
		os.Exit(1)
	}
}

