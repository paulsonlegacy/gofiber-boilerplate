package storage

import (
	"context"
	"fmt"
	"io"

	"boilerplate/internal/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryStorage implements the Storage interface using Cloudinary.
type CloudinaryStorage struct {
	client *cloudinary.Cloudinary
}

// NewCloudinaryStorage creates a Cloudinary-backed storage provider from config.
// Returns a no-op stub if credentials are missing — safe for local development.
func NewCloudinaryStorage(cfg config.CloudinaryConfig) *CloudinaryStorage {
	if cfg.CloudName == "" || cfg.APIKey == "" || cfg.APISecret == "" {
		return &CloudinaryStorage{} // graceful no-op in dev
	}
	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return &CloudinaryStorage{}
	}
	return &CloudinaryStorage{client: cld}
}

func (s *CloudinaryStorage) UploadFile(ctx context.Context, filePath, folder string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("storage not configured")
	}
	result, err := s.client.Upload.Upload(ctx, filePath, uploader.UploadParams{Folder: folder})
	if err != nil {
		return "", err
	}
	return result.SecureURL, nil
}

func (s *CloudinaryStorage) UploadBytes(ctx context.Context, data []byte, fileName, folder string) (string, string, error) {
	if s.client == nil {
		return "", "", fmt.Errorf("storage not configured")
	}
	result, err := s.client.Upload.Upload(ctx, data, uploader.UploadParams{
		PublicID: fileName,
		Folder:   folder,
	})
	if err != nil {
		return "", "", err
	}
	return result.SecureURL, result.PublicID, nil
}

func (s *CloudinaryStorage) UploadReader(ctx context.Context, r io.Reader, fileName, folder string) (string, string, error) {
	if s.client == nil {
		return "", "", fmt.Errorf("storage not configured")
	}
	result, err := s.client.Upload.Upload(ctx, r, uploader.UploadParams{
		PublicID: fileName,
		Folder:   folder,
	})
	if err != nil {
		return "", "", err
	}
	return result.SecureURL, result.PublicID, nil
}

func (s *CloudinaryStorage) DeleteFile(ctx context.Context, publicID string) error {
	if s.client == nil {
		return nil
	}
	_, err := s.client.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}

func (s *CloudinaryStorage) DeleteFolder(ctx context.Context, folderPath string) error {
	if s.client == nil {
		return nil
	}
	_, err := s.client.Admin.DeleteFolder(ctx, admin.DeleteFolderParams{Folder: folderPath})
	return err
}

func (s *CloudinaryStorage) GetFileURL(folder, publicID, resourceType string) string {
	if s.client == nil {
		return ""
	}
	return fmt.Sprintf("https://res.cloudinary.com/%s/%s/upload/%s/%s", folder, resourceType, folder, publicID)
}

// ensure interface compliance
var _ Storage = (*CloudinaryStorage)(nil)
