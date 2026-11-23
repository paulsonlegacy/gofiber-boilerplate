// Package storage provides a file storage abstraction.
// All file operations go through the Storage interface — swap Cloudinary for S3, GCS,
// or local disk by implementing the interface and injecting the new provider at startup.
package storage

import (
	"context"
	"io"
)

// Storage is the contract all file storage backends must satisfy.
type Storage interface {
	UploadFile(ctx context.Context, filePath string, folder string) (url string, err error)
	UploadBytes(ctx context.Context, data []byte, fileName, folder string) (url, publicID string, err error)
	UploadReader(ctx context.Context, r io.Reader, fileName, folder string) (url, publicID string, err error)
	DeleteFile(ctx context.Context, publicID string) error
	DeleteFolder(ctx context.Context, folderPath string) error
	GetFileURL(folder, publicID, resourceType string) string
}
