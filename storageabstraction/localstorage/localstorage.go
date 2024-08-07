package localstorage

import (
	"fmt"
	"github.com/2flow/gokies/storageabstraction"
	"github.com/2flow/gokies/storageabstraction/common"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
)

type localStorage struct {
	storageabstraction.IFileStorage
	rootDirectory string
}

// NewLocalStorage creates a new instance of an local storage
func NewLocalStorage(rootDir string) storageabstraction.IFileStorage {
	err := os.MkdirAll(rootDir, 0777|os.ModeDir)
	if err != nil {
		return nil
	}
	return &localStorage{rootDirectory: rootDir}
}

// create path if not exists, and set the owner of it
func createFolder(path string, uid, gid int) {
	path = filepath.ToSlash(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		subPath, filepath := filepath.Split(path)
		if filepath == "" || subPath == "" {
			return
		}

		createFolder(subPath, uid, gid)

		err := os.MkdirAll(path, 0777|os.ModeDir)
		if err != nil {
			_ = fmt.Errorf("[createFolder] Error during mkdir %s", err.Error())
		}
		err = os.Chown(path, uid, gid)
		if err != nil {
			_ = fmt.Errorf("[createFolder] Error during chown %s", err.Error())
		}
	}
}

func (storage *localStorage) Write(fileName string, _ int64, reader io.ReadSeeker) error {
	filePath := path.Join(storage.rootDirectory, fileName)
	filePath = filepath.ToSlash(filePath)

	dirPath, _ := filepath.Split(filePath)
	if len(dirPath) < len(storage.rootDirectory) {
		return nil
	}

	group, err := user.Lookup("www-data")
	if err != nil {
		_ = fmt.Errorf("[LocalStorageWrite]" + "Unable to find group www-data")
		return err
	}
	uid, _ := strconv.Atoi(group.Uid)
	gid, _ := strconv.Atoi(group.Gid)
	createFolder(dirPath, uid, gid)

	/*if err != nil {
		fmt.Errorf("[LocalStorageWrite]"+"Unable create directory %s, FILE: %s, ERROR: %s", dirPath, filePath,
			err.Error())
		return err
	}*/

	err = os.Remove(filePath)
	if err != nil {
		_ = fmt.Errorf("[LocalStorageWrite]"+"Unable to remove file %s: %s", filePath, err.Error())
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0777|os.ModeDir)

	if err != nil {
		_ = fmt.Errorf("[LocalStorageWrite]"+"Unable to Open file %s: %s", filePath, err.Error())
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
			errStr := "actual write Count is not equal to bytesCount"
			if err != nil {
				errStr = err.Error()
			}

			_ = fmt.Errorf("[LocalStorageWrite]"+"Unable to write file %s: %s", filePath, errStr)
			return err
		}

	}

	return nil
}

func (storage *localStorage) Read(fileName string) (io.ReadCloser, error) {
	return os.OpenFile(path.Join(storage.rootDirectory, fileName), os.O_RDONLY, 0644)
}

func (storage *localStorage) Join(paths ...string) string {
	return common.LinuxPathJoin(paths...)
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
	rootPath = filepath.ToSlash(rootPath)
	trimLen := len(rootPath)
	if directory[len(directory)-1] != '/' && directory[len(directory)-1] != '\\' {
		trimLen++
	}

	return filepath.Walk(path.Join(storage.rootDirectory, directory), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			_ = walk("", info, err)
			return err
		}
		path = filepath.ToSlash(path)

		if len(path) <= trimLen {
			path = ""
		} else {
			path = path[trimLen:]
		}

		return walk(path, info, nil)
	})
}
