package scanner

import (
	"bytes"
	"compress/zlib"
	"io"
	"os"
	"regexp"
	"strings"
)

func processPdfFile(pdfPath string, ignoredDomains []string, remove, dryRun bool) []string {
	file, err := os.Open(pdfPath)
	if err != nil {
		Logger.Errorf("Error opening PDF file %s: %v", pdfPath, err)
		return nil
	}
	defer file.Close()

	pdfContent, err := io.ReadAll(file)
	if err != nil {
		Logger.Errorf("Error reading PDF file %s: %v", pdfPath, err)
		return nil
	}

	streams := regexp.MustCompile(`stream[\r\n\s]+(.*?)[\r\n\s]+endstream`).FindAllSubmatch(pdfContent, -1)
	var foundUrls []string
	for _, stream := range streams {
		urls := extractUrlsFromPdfStream(stream[1], ignoredDomains)
		foundUrls = append(foundUrls, urls...)
	}

	if len(foundUrls) > 0 && remove && !dryRun {
		// backup
		if err := backupFile(pdfPath); err != nil {
			Logger.Errorf("Error creating backup for %s: %v", pdfPath, err)
		}
		modifiedContent := removeURLsFromContent(pdfContent, foundUrls)
		err = os.WriteFile(pdfPath, modifiedContent, 0644)
		if err != nil {
			Logger.Errorf("Error writing PDF file %s: %v", pdfPath, err)
		}
	}

	return foundUrls
}

func extractUrlsFromPdfStream(stream []byte, ignored []string) []string {
	var foundUrls []string
	b := bytes.NewReader(stream)
	r, err := zlib.NewReader(b)
	if err != nil {
		return foundUrls
	}
	defer r.Close()

	decompressedData, err := io.ReadAll(r)
	if err != nil {
		return foundUrls
	}

	urlPattern := regexp.MustCompile(`https?://[^\s<>"'{}|\\^` + "`" + `]+`)
	matches := urlPattern.FindAll(decompressedData, -1)
	for _, match := range matches {
		urlStr := string(match)
		if isIgnoredDomain(urlStr, ignored) {
			continue
		}
		foundUrls = append(foundUrls, urlStr)
	}

	return foundUrls
}

func isIgnoredDomain(url string, ignored []string) bool {
	for _, domain := range ignored {
		if strings.Contains(url, domain) {
			return true
		}
	}
	return false
}
