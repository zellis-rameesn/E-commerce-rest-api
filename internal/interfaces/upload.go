package interfaces

import "mime/multipart"

type UploadInterface interface {
	UploadFile(file *multipart.FileHeader, path string) (string, error)
	DeleteFile(path string) error
}
