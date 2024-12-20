package report

type SummaryData struct {
	TotalFiles      int `json:"total_files"`
	SuspiciousFiles int `json:"suspicious_files"`
	NormalFiles     int `json:"normal_files"`
}
