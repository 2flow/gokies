package localstorage

import (
	"github.com/2flow/gokies/storageabstraction"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type localStorage struct {
	storageabstraction.IFileStorage
	rootDirectory string
}

// NewLocalStorage creates a new instance of an local storage
func NewLocalStorage(rootDir string) storageabstraction.IFileStorage {
	return &localStorage{rootDirectory: rootDir}
}

func (storage *localStorage) UploadFile(fileName string, _ int64, reader io.ReadSeeker) error {
	file, err := os.OpenFile(path.Join(storage.rootDirectory, fileName), os.O_CREATE, 0644)

	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024)

	for {
		bytesCount, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		writeCount, err := file.Write(buffer[:bytesCount])
		if err != nil || writeCount != bytesCount {
			return err
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

func (storage *localStorage) DownloadFile(fileName string) (io.ReadCloser, error) {
	return os.OpenFile(path.Join(storage.rootDirectory, fileName), os.O_RDONLY, 0644)
}

func (storage *localStorage) FileSize(fileName string) (int64, error) {
	stats, err := os.Stat(path.Join(storage.rootDirectory, fileName))
	if err != nil {
		return 0, err
	}

	return stats.Size(), nil
}

func (storage *localStorage) DeleteDirectory(directory string) error {
	return os.RemoveAll(path.Join(storage.rootDirectory, directory))
}

func (storage *localStorage) DeleteFile(fileName string) error {
	return os.Remove(path.Join(storage.rootDirectory, fileName))
}

func (storage *localStorage) Walk(directory string, walk storageabstraction.WalkFunc) error {
	return filepath.Walk(path.Join(storage.rootDirectory, directory), func(path string, info fs.FileInfo, err error) error {

		if err != nil {
			_ = walk("", storageabstraction.FileInfo{}, err)
			return err
		}

		return walk(strings.TrimPrefix(path, directory), storageabstraction.FileInfo{
			Size:  info.Size(),
			IsDir: info.IsDir(),
		}, nil)
	})
}
