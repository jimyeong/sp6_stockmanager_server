package models

type ImageDeleteRequest struct {
	ImagePath string `json:"image_path"`
}

type ImageDeleteResponse struct {
	Message   string `json:"message"`
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
	ImagePath string `json:"image_path"`
}
type ImageUploadResponse struct {
	ImagePath string `json:"image_path"`
	ImageID   string `json:"image_id"`
	FileSize  int64  `json:"file_size"`
	Message   string `json:"message"`
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
	FileName  string `json:"file_name"`
}
