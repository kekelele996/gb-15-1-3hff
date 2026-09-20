package dto

// PlagiarismResponse 查重结果响应。
type PlagiarismResponse struct {
	ID         uint    `json:"id"`
	PaperID    uint    `json:"paper_id"`
	Similarity float64 `json:"similarity"`
	Status     string  `json:"status"`
	Report     string  `json:"report"`
	CheckedAt  string  `json:"checked_at"`
}
