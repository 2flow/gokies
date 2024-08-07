package storageabstraction

import (
	"io"
	"io/fs"
	"os"
	"time"
)

type WalkFunc func(path string, info os.FileInfo, err error) error

type FileInfo struct {
	size  int64
	isDir bool
}

func NewFileInfo(size int64, isDir bool) *FileInfo {
	return &FileInfo{
		size:  size,
		isDir: isDir,
	}
}

// IFileStorage is the interface for a filestorage used by the ecosystem
type IFileStorage interface {
	// CreateDirectory(dirName string) error

	Write(fileName string, fileSize int64, reader io.ReadSeeker) error
	Read(fileName string) (io.ReadCloser, error)
	FileSize(fileName string) (int64, error)
	DeleteDirectory(directory string) error
	DeleteFile(fileName string) error
	Walk(directory string, walk WalkFunc) error
	// Join the paths together and clean up path
	Join(paths ...string) string
}

func (fileInfo *FileInfo) Name() string {
	return ""
}

func (fileInfo *FileInfo) Size() int64 {
	return fileInfo.size
}

func (fileInfo *FileInfo) Mode() fs.FileMode {
	return fs.ModeIrregular
}

func (fileInfo *FileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fileInfo *FileInfo) IsDir() bool {
	return fileInfo.isDir
}

func (fileInfo *FileInfo) Sys() any {
	return nil
}
