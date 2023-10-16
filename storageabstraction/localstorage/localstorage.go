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

func (storage *localStorage) Write(fileName string, _ int64, reader io.ReadSeeker) error {
	filePath := path.Join(storage.rootDirectory, fileName)
	filePath = filepath.ToSlash(filePath)

	pathEndIdx := strings.LastIndex(filePath, "/")
	err := os.MkdirAll(filePath[:pathEndIdx], 0777)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE, 0644)

	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024)

	for {
		bytesCount, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			break
		}

		writeCount, err := file.Write(buffer[:bytesCount])
		if err != nil || writeCount != bytesCount {
			return err
		}

	}

	return nil
}

func (storage *localStorage) Read(fileName string) (io.ReadCloser, error) {
	return os.OpenFile(path.Join(storage.rootDirectory, fileName), os.O_RDONLY, 0644)
}

func (storage *localStorage) Join(paths ...string) string {
	joinedPath := ""
	for _, path := range paths {

		if joinedPath == "" {
			joinedPath = path
			continue
		}
		if joinedPath[len(joinedPath)-1] != '/' && joinedPath[len(joinedPath)-1] != '\\' {
			joinedPath += "/"
		}

		if path[0] == '/' || path[0] == '\\' {
			joinedPath += path[1:]
		} else {
			joinedPath += path
		}
	}

	return joinedPath
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

	rootPath := storage.Join(storage.rootDirectory, directory)
	trimLen := len(rootPath)
	if directory[len(directory)-1] != '/' && directory[len(directory)-1] != '\\' {
		trimLen++
	}

	return filepath.Walk(path.Join(storage.rootDirectory, directory), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			_ = walk("", info, err)
			return err
		}

		if len(path) <= trimLen {
			path = ""
		} else {
			path = path[trimLen:]
		}

		return walk(path, info, nil)
	})
}
