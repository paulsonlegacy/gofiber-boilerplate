package util

import (
	"fmt"
	"mime/multipart"
	"strings"
)

const maxDocumentSize = 10 * 1024 * 1024 // 10 MB

var allowedDocumentExtensions = map[string]bool{
	".pdf":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// ValidateDocumentFile checks that a file is one of the allowed types and under the size limit.
// Enforce this on all user-uploaded documents at the handler layer.
func ValidateDocumentFile(fh *multipart.FileHeader) error {
	if fh == nil {
		return fmt.Errorf("no file provided")
	}
	if fh.Size > maxDocumentSize {
		return fmt.Errorf("file too large: maximum allowed size is 10MB")
	}

	name := strings.ToLower(fh.Filename)
	for ext := range allowedDocumentExtensions {
		if strings.HasSuffix(name, ext) {
			return nil
		}
	}
	return fmt.Errorf("unsupported file type: allowed types are PDF, JPG, JPEG, PNG")
}

// ValidateImageFile checks that a file is a supported image type and under 5 MB.
func ValidateImageFile(fh *multipart.FileHeader) error {
	if fh == nil {
		return fmt.Errorf("no file provided")
	}
	const maxImageSize = 5 * 1024 * 1024
	if fh.Size > maxImageSize {
		return fmt.Errorf("image too large: maximum allowed size is 5MB")
	}

	name := strings.ToLower(fh.Filename)
	allowed := []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}
	for _, ext := range allowed {
		if strings.HasSuffix(name, ext) {
			return nil
		}
	}
	return fmt.Errorf("unsupported image type: allowed types are JPG, PNG, WEBP, GIF")
}
