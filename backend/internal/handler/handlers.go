package handler

// Handlers 聚合全部 HTTP 处理器。
type Handlers struct {
	Auth       *AuthHandler
	User       *UserHandler
	Paper      *PaperHandler
	Review     *ReviewHandler
	Revision   *RevisionHandler
	Plagiarism *PlagiarismHandler
	Audit      *AuditHandler
	Statistics *StatisticsHandler
	File       *FileHandler
	Health     *HealthHandler
}
