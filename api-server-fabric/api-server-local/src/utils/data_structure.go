package utils

type SourceInfo struct {
	SourceID string `json:"sourceID"`
	SourceName string `json:"sourceName"`
	SourceURL string `json:"sourceURL"`
	SourceCredits float64 `json:"sourceCredits"`
	Reserved string `json:"reserved"`
}

type DragLog struct {
	LogID string `json:"logID"`
	Input string `json:"from"`
	Output string `json:"to"`
	Reserved string `json:"reserved"`
}

type DragLogList struct {
	Logs []DragLog `json:"logs"`
	
}