package scanner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Lercas/SneakPeeker/report"
)

func ScanPath(path string, ignoredDomains []string, remove bool, dryRun bool, verbose bool, workers int) ([]report.Report, report.SummaryData, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, report.SummaryData{}, fmt.Errorf("path %s does not exist", path)
	}

	var files []string
	if info.IsDir() {
		err := filepath.Walk(path, func(p string, i os.FileInfo, e error) error {
			if e != nil {
				return e
			}
			if !i.IsDir() {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return nil, report.SummaryData{}, err
		}
	} else {
		files = append(files, path)
	}

	jobs := make(chan string, len(files))
	results := make(chan report.Report, len(files))

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				r := processFile(f, ignoredDomains, remove, dryRun, verbose)
				results <- r
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	wg.Wait()
	close(results)

	var reports []report.Report
	var summary report.SummaryData
	for r := range results {
		reports = append(reports, r)
		if r.Suspicious {
			summary.SuspiciousFiles++
		} else {
			summary.NormalFiles++
		}
	}
	summary.TotalFiles = len(reports)

	return reports, summary, nil
}

func processFile(filePath string, ignoredDomains []string, remove, dryRun, verbose bool) report.Report {
	fi, err := os.Stat(filePath)
	if err != nil {
		Logger.Errorf("Error stating file %s: %v", filePath, err)
		return report.Report{FilePath: filePath}
	}

	foundUrls := scanFile(filePath, ignoredDomains, remove, dryRun)
	suspicious := len(foundUrls) > 0

	if suspicious {
		Logger.Infof("The file %s is suspicious. URLs found: %v", filePath, foundUrls)
	} else {
		Logger.Debugf("The file %s seems normal.", filePath)
	}

	return report.Report{
		FilePath:    filePath,
		Suspicious:  suspicious,
		FoundURLs:   foundUrls,
		FileSize:    fi.Size(),
		ProcessedAt: time.Now().Format(time.RFC3339),
	}
}

func scanFile(filePath string, ignoredDomains []string, remove, dryRun bool) []string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".pdf":
		return processPdfFile(filePath, ignoredDomains, remove, dryRun)
	case ".zip", ".docx", ".xlsx", ".pptx":
		return decompressAndScan(filePath, ignoredDomains, remove, dryRun)
	case ".txt", ".html":
		return processTextFile(filePath, ignoredDomains, remove, dryRun)
	default:
		return []string{}
	}
}

func backupFile(filePath string) error {
	backupPath := filePath + ".bak"
	input, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	return err
}
