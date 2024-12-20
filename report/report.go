package report

type FinalReport struct {
	Reports []Report    `json:"reports"`
	Summary SummaryData `json:"summary"`
}

type Report struct {
	FilePath    string   `json:"file_path"`
	Suspicious  bool     `json:"suspicious"`
	FoundURLs   []string `json:"found_urls"`
	FileSize    int64    `json:"file_size"`
	ProcessedAt string   `json:"processed_at"`
}
