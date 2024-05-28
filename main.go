package main

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Report struct {
	FilePath   string   `json:"file_path"`
	Suspicious bool     `json:"suspicious"`
	FoundURLs  []string `json:"found_urls"`
}

var (
	removeCanary = flag.Bool("f", false, "Remove canary tokens from files")
	reportFile   = flag.String("r", "report.json", "Report file name")
)

func main() {
	printBanner()
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Usage: go run script.go [-f] [-r report_file] FILE_OR_DIRECTORY_PATH")
		return
	}

	path := flag.Args()[0]
	var reports []Report

	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					report := processFile(path)
					reports = append(reports, report)
				}
				return nil
			})
			if err != nil {
				fmt.Printf("[ERROR] Error walking the path %s: %v\n", path, err)
			}
		} else {
			report := processFile(path)
			reports = append(reports, report)
		}
	} else {
		fmt.Printf("[ERROR] The path %s does not exist.\n", path)
	}

	reportData, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		fmt.Printf("[ERROR] Error marshalling report data: %v\n", err)
		return
	}

	err = os.WriteFile(*reportFile, reportData, 0644)
	if err != nil {
		fmt.Printf("[ERROR] Error writing report file: %v\n", err)
	} else {
		fmt.Printf("\n[INFO] Report written to %s\n", *reportFile)
	}
}

func printBanner() {
	banner := `                             
 _____             _   _____         _           
|   __|___ ___ ___| |_|  _  |___ ___| |_ ___ ___ 
|__   |   | -_| .'| '_|   __| -_| -_| '_| -_|  _|
|_____|_|_|___|__,|_,_|__|  |___|___|_,_|___|_|  
												 
By @belka_e
`
	fmt.Println(banner)
}

func processFile(filePath string) Report {
	var foundUrls []string
	var suspicious bool
	if filepath.Ext(filePath) == ".pdf" {
		foundUrls = processPdfFile(filePath)
	} else if filepath.Ext(filePath) == ".zip" || filepath.Ext(filePath) == ".docx" || filepath.Ext(filePath) == ".xlsx" || filepath.Ext(filePath) == ".pptx" {
		foundUrls = decompressAndScan(filePath)
	}
	suspicious = len(foundUrls) > 0

	if suspicious {
		fmt.Printf("\n[INFO] The file %s is suspicious. URLs found:\n", filePath)
		for _, url := range foundUrls {
			fmt.Println(url)
		}
		if *removeCanary {
			removeCanaryTokens(filePath)
		}
	} else {
		fmt.Printf("\n[INFO] The file %s seems normal.\n", filePath)
	}

	return Report{
		FilePath:   filePath,
		Suspicious: suspicious,
		FoundURLs:  foundUrls,
	}
}

func processPdfFile(pdfPath string) []string {
	file, err := os.Open(pdfPath)
	if err != nil {
		fmt.Printf("[ERROR] Error opening PDF file %s: %v\n", pdfPath, err)
		return nil
	}
	defer file.Close()

	pdfContent, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("[ERROR] Error reading PDF file %s: %v\n", pdfPath, err)
		return nil
	}

	streams := regexp.MustCompile(`stream[\r\n\s]+(.*?)[\r\n\s]+endstream`).FindAllSubmatch(pdfContent, -1)
	var foundUrls []string
	for _, stream := range streams {
		urls := extractUrlsFromStream(stream[1])
		foundUrls = append(foundUrls, urls...)
	}

	if *removeCanary {
		modifiedContent := removeUrlsFromPdfContent(pdfContent, foundUrls)
		err = os.WriteFile(pdfPath, modifiedContent, 0644)
		if err != nil {
			fmt.Printf("[ERROR] Error writing PDF file %s: %v\n", pdfPath, err)
		}
	}

	return foundUrls
}

func extractUrlsFromStream(stream []byte) []string {
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
		foundUrls = append(foundUrls, string(match))
	}

	return foundUrls
}

func removeUrlsFromPdfContent(content []byte, urls []string) []byte {
	modifiedContent := string(content)
	for _, url := range urls {
		modifiedContent = strings.ReplaceAll(modifiedContent, url, "")
	}
	return []byte(modifiedContent)
}

func decompressAndScan(filePath string) []string {
	var foundUrls []string
	fileDir := filepath.Dir(filePath)
	tempDir := filepath.Join(fileDir, "temp_extracted")

	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		fmt.Printf("[ERROR] Error creating temporary directory: %v\n", err)
		return nil
	}
	defer os.RemoveAll(tempDir)

	r, err := zip.OpenReader(filePath)
	if err != nil {
		fmt.Printf("[ERROR] Error opening zip file %s: %v\n", filePath, err)
		return nil
	}
	defer r.Close()

	urlPattern := regexp.MustCompile(`https?://\S+`)
	ignoredDomains := []string{"schemas.openxmlformats.org", "schemas.microsoft.com", "purl.org", "w3.org"}

	for _, file := range r.File {
		extractedFilePath := filepath.Join(tempDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(extractedFilePath, os.ModePerm)
			continue
		}

		err := os.MkdirAll(filepath.Dir(extractedFilePath), os.ModePerm)
		if err != nil {
			fmt.Printf("[ERROR] Error creating directory for file %s: %v\n", extractedFilePath, err)
			continue
		}

		rc, err := file.Open()
		if err != nil {
			fmt.Printf("[ERROR] Error opening file %s in zip: %v\n", file.Name, err)
			continue
		}

		outFile, err := os.Create(extractedFilePath)
		if err != nil {
			fmt.Printf("[ERROR] Error creating file %s: %v\n", extractedFilePath, err)
			rc.Close()
			continue
		}

		_, err = io.Copy(outFile, rc)
		if err != nil {
			fmt.Printf("[ERROR] Error extracting file %s: %v\n", extractedFilePath, err)
			outFile.Close()
			rc.Close()
			continue
		}

		outFile.Close()
		rc.Close()

		content, err := os.ReadFile(extractedFilePath)
		if err != nil {
			fmt.Printf("[ERROR] Error reading file %s: %v\n", extractedFilePath, err)
			continue
		}

		urls := urlPattern.FindAll(content, -1)
		for _, url := range urls {
			urlStr := string(url)
			isIgnored := false
			for _, domain := range ignoredDomains {
				if strings.Contains(urlStr, domain) {
					isIgnored = true
					break
				}
			}
			if !isIgnored {
				foundUrls = append(foundUrls, urlStr)
				if *removeCanary {
					content = []byte(strings.ReplaceAll(string(content), urlStr, ""))
					if err := os.WriteFile(extractedFilePath, content, 0644); err != nil {
						fmt.Printf("[ERROR] Error writing to file %s: %v\n", extractedFilePath, err)
					}
				}
			}
		}
	}

	if *removeCanary {
		err := createNewZip(filePath, tempDir)
		if err != nil {
			fmt.Printf("[ERROR] Error creating new zip file %s: %v\n", filePath, err)
		}
	}

	return foundUrls
}

func removeCanaryTokens(filePath string) {
	if filepath.Ext(filePath) == ".pdf" {
		removeCanaryTokensFromPdf(filePath)
	} else if filepath.Ext(filePath) == ".zip" || filepath.Ext(filePath) == ".docx" || filepath.Ext(filePath) == ".xlsx" || filepath.Ext(filePath) == ".pptx" {
		removeCanaryTokensFromArchive(filePath)
	}
}

func removeCanaryTokensFromPdf(filePath string) {
	fmt.Printf("[INFO] Removing canary tokens from PDF file: %s\n", filePath)
	foundUrls := processPdfFile(filePath)
	if len(foundUrls) > 0 {
		input, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("[ERROR] Error reading PDF file %s: %v\n", filePath, err)
			return
		}

		modifiedContent := removeUrlsFromPdfContent(input, foundUrls)
		err = os.WriteFile(filePath, modifiedContent, 0644)
		if err != nil {
			fmt.Printf("[ERROR] Error writing PDF file %s: %v\n", filePath, err)
		}
	}
}

func removeCanaryTokensFromArchive(filePath string) {
	fmt.Printf("[INFO] Removing canary tokens from archive file: %s\n", filePath)
	decompressAndScan(filePath)
}

func createNewZip(filePath, tempDir string) error {
	zipFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
