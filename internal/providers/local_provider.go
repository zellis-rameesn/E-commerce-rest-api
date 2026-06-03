package providers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalProvider struct {
	basePath string
}

func NewLocalProvider(path string) *LocalProvider {
	return &LocalProvider{
		basePath: path,
	}
}

func (p *LocalProvider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	fullPath := filepath.Join(p.basePath, path)

	// create directory
	// fullpath contains the filename hence using filepath.dir to cut out the filename
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}

	// Open source
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// create dest
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// read from source to destination
	if _, err := dst.ReadFrom(src); err != nil {
		return "", err
	}
	return fmt.Sprintf("/uploads/%s", path), nil
}

func (p *LocalProvider) DeleteFile(path string) error {
	fullPath := filepath.Join(p.basePath, path)
	return os.Remove(fullPath)
}
