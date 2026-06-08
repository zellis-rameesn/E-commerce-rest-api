package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/zellis-rameesn/go-ecommerce/internal/interfaces"
)

type UploadService struct {
	Provider interfaces.UploadInterface
}

func NewUploadService(provider interfaces.UploadInterface) *UploadService {
	return &UploadService{
		Provider: provider,
	}
}

func (u *UploadService) UploadProductImage(productID uint, file *multipart.FileHeader) (string, error) {

	ext := strings.ToLower(filepath.Ext(file.Filename))

	if !isValidImageExt(ext) {
		return "", fmt.Errorf("Invalid file type %s", ext)
	}

	path := fmt.Sprintf("products/%d/%s%s", productID, uuid.New().String(), ext)

	return u.Provider.UploadFile(file, path)
}

func isValidImageExt(ext string) bool {
	allValidExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif"}

	for _, validExt := range allValidExts {
		if ext == validExt {
			return true
		}
	}
	return false
}
