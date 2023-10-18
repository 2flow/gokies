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

func (fileManager FileManager) GetUploadWriter(path string) (io.WriteCloser, error) {
	return fileManager.uploader.UploadTar(path)
}

func (fileManager FileManager) UploadTar(path string, reader io.Reader) error {
	uploader, err := fileManager.GetUploadWriter(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(uploader, reader)
	if err != nil {
		return err
	}

	_ = uploader.Close()

	return nil
}

func (fileManager FileManager) BackupDirectory(path string, writer io.Writer) error {
	compressor := compression.NewCompression(fileManager.storage)
	return compressor.CompressDir(path, writer)
}
