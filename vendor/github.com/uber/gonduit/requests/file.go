package requests

// FileDownloadRequest represents a call to file.download.
type FileDownloadRequest struct {
	PHID string `json:"phid"`
	Request
}
