package storageabstraction

import "io"

// IFileStorage is the interface for a filestorage used by the ecosystem
type IFileStorage interface {
	// CreateDirectory(dirName string) error

	UploadFile(fileName string, fileSize int64, reader io.ReadSeeker) error
	DownloadFile(fileName string) (io.ReadCloser, error)
	FileSize(fileName string) (int64, error)
	DeleteDirectory(directory string) error
}
