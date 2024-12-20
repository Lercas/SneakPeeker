package scanner

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func decompressAndScan(filePath string, ignoredDomains []string, remove, dryRun bool) []string {
	fileDir := filepath.Dir(filePath)
	tempDir := filepath.Join(fileDir, "temp_extracted")

	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		Logger.Errorf("Error creating temporary directory: %v", err)
		return nil
	}
	defer os.RemoveAll(tempDir)

	r, err := zip.OpenReader(filePath)
	if err != nil {
		Logger.Errorf("Error opening zip file %s: %v", filePath, err)
		return nil
	}
	defer r.Close()

	urlPattern := regexp.MustCompile(`https?://\S+`)

	var foundUrls []string

	for _, file := range r.File {
		extractedFilePath := filepath.Join(tempDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(extractedFilePath, os.ModePerm)
			continue
		}

		err := os.MkdirAll(filepath.Dir(extractedFilePath), os.ModePerm)
		if err != nil {
			Logger.Errorf("Error creating directory for file %s: %v", extractedFilePath, err)
			continue
		}

		rc, err := file.Open()
		if err != nil {
			Logger.Errorf("Error opening file %s in zip: %v", file.Name, err)
			continue
		}

		outFile, err := os.Create(extractedFilePath)
		if err != nil {
			Logger.Errorf("Error creating file %s: %v", extractedFilePath, err)
			rc.Close()
			continue
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			Logger.Errorf("Error extracting file %s: %v", extractedFilePath, err)
			continue
		}

		content, err := os.ReadFile(extractedFilePath)
		if err != nil {
			Logger.Errorf("Error reading file %s: %v", extractedFilePath, err)
			continue
		}

		urls := urlPattern.FindAll(content, -1)
		var fileModified bool
		for _, u := range urls {
			urlStr := string(u)
			if !isIgnoredDomain(urlStr, ignoredDomains) {
				foundUrls = append(foundUrls, urlStr)
				if remove && !dryRun {
					content = []byte(strings.ReplaceAll(string(content), urlStr, ""))
					fileModified = true
				}
			}
		}

		if fileModified {
			if err := os.WriteFile(extractedFilePath, content, 0644); err != nil {
				Logger.Errorf("Error writing to file %s: %v", extractedFilePath, err)
			}
		}
	}

	if len(foundUrls) > 0 && remove && !dryRun {
		// backup
		if err := backupFile(filePath); err != nil {
			Logger.Errorf("Error creating backup for %s: %v", filePath, err)
		}
		err := createNewZip(filePath, tempDir)
		if err != nil {
			Logger.Errorf("Error creating new zip file %s: %v", filePath, err)
		}
	}

	return foundUrls
}

func createNewZip(filePath, tempDir string) error {
	zipFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}

		writer, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return nil
}
