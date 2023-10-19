package filecontainer

import (
	"github.com/2flow/gokies/compression"
	"github.com/2flow/gokies/storageabstraction"
	"github.com/go-kit/log"
	"io"
)

type FileManager struct {
	storage  storageabstraction.IFileStorage
	logger   log.Logger
	uploader *Uploader
}

func CreateFileManager(storage storageabstraction.IFileStorage, logger log.Logger) *FileManager {
	return &FileManager{storage: storage,
		logger:   logger,
		uploader: CreateUploader("./temps/", logger, storage)}
}

// GetFile returns an error or a stream of data which represent the requested file
func (fileManager FileManager) GetFile(path string) (io.ReadCloser, error) {
	reader, err := fileManager.storage.Read(path)
	if err != nil {
		fileManager.logger.Log("msg", "Unable to read from storage", "error", err.Error())
		return nil, err
	}

	return reader, nil
}

/*
func (fileManager FileManager) DoesFileExist(path string) {

}
*/

func (fileManager FileManager) GetUploadWriter(path string, callbacks UploadCallBacks) (TarUploader, error) {
	return fileManager.uploader.UploadTar(path, callbacks)
}

func (fileManager FileManager) UploadTar(path string, callbacks UploadCallBacks, reader io.Reader) error {
	uploader, err := fileManager.GetUploadWriter(path, callbacks)
	if err != nil {
		return err
	}

	_, err = io.Copy(uploader, reader)
	if err != nil {
		uploader.Error()
		return err
	}

	uploader.Done()

	return nil
}

func (fileManager FileManager) BackupDirectory(path string, writer io.Writer) error {
	compressor := compression.NewCompression(fileManager.storage)
	return compressor.CompressDir(path, writer)
}
