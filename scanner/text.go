package scanner

import (
	"os"
	"regexp"
	"strings"
)

func processTextFile(filePath string, ignoredDomains []string, remove, dryRun bool) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		Logger.Errorf("Error reading text file %s: %v", filePath, err)
		return nil
	}

	urlPattern := regexp.MustCompile(`https?://\S+`)
	matches := urlPattern.FindAll(content, -1)
	var foundUrls []string
	modifiedContent := content

	for _, m := range matches {
		urlStr := string(m)
		if isIgnoredDomain(urlStr, ignoredDomains) {
			continue
		}
		foundUrls = append(foundUrls, urlStr)
		if remove && !dryRun {
			modifiedContent = []byte(strings.ReplaceAll(string(modifiedContent), urlStr, ""))
		}
	}

	if len(foundUrls) > 0 && remove && !dryRun {
		if err := backupFile(filePath); err != nil {
			Logger.Errorf("Error creating backup for %s: %v", filePath, err)
		}
		if err := os.WriteFile(filePath, modifiedContent, 0644); err != nil {
			Logger.Errorf("Error writing text file %s: %v", filePath, err)
		}
	}

	return foundUrls
}
