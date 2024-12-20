package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Lercas/SneakPeeker/report"
	"github.com/Lercas/SneakPeeker/scanner"
)

var (
	removeCanary  = flag.Bool("f", false, "Remove canary tokens from files")
	reportFile    = flag.String("r", "report.json", "Report file name")
	verbose       = flag.Bool("v", false, "Enable verbose output")
	dryRun        = flag.Bool("dry-run", false, "Do not modify any files, just report")
	ignoreDomains = flag.String("ignore", "", "Comma-separated list of domains to ignore")
	workersNum    = flag.Int("w", 5, "Number of workers for parallel scan")
)

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

func main() {
	printBanner()
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Usage: go run main.go [options] FILE_OR_DIRECTORY_PATH")
		flag.PrintDefaults()
		return
	}

	scanner.InitLogger(*verbose)

	path := flag.Args()[0]
	var ignoredDomains []string
	if *ignoreDomains != "" {
		ignoredDomains = strings.Split(*ignoreDomains, ",")
		for i, d := range ignoredDomains {
			ignoredDomains[i] = strings.TrimSpace(d)
		}
	}

	start := time.Now()

	reports, summary, err := scanner.ScanPath(path, ignoredDomains, *removeCanary, *dryRun, *verbose, *workersNum)
	if err != nil {
		scanner.Logger.Errorf("Error scanning path: %v", err)
		return
	}

	finalReport := report.FinalReport{
		Reports: reports,
		Summary: summary,
	}

	reportData, err := json.MarshalIndent(finalReport, "", "  ")
	if err != nil {
		scanner.Logger.Errorf("Error marshalling report data: %v", err)
		return
	}

	err = os.WriteFile(*reportFile, reportData, 0644)
	if err != nil {
		scanner.Logger.Errorf("Error writing report file: %v", err)
	} else {
		scanner.Logger.Infof("Report written to %s", *reportFile)
	}

	scanner.Logger.Infof("Scanning completed in %s", time.Since(start))
}
