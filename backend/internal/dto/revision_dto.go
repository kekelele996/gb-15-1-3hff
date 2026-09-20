package dto

// RevisionResponse 修稿记录响应。
type RevisionResponse struct {
	ID             uint   `json:"id"`
	PaperID        uint   `json:"paper_id"`
	Version        int    `json:"version"`
	FileName       string `json:"file_name"`
	FileKey        string `json:"file_key"`
	ResponseLetter string `json:"response_letter"`
	SubmittedByID  uint   `json:"submitted_by_id"`
	CreatedAt      string `json:"created_at"`
}
