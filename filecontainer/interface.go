package filecontainer

import "io"

type IFileManager interface {
	GetFile(path string) (io.ReadCloser, error)
	GetUploadWriter(path string, callbacks UploadCallBacks) (TarUploader, error)
	BackupDirectory(path string, writer io.Writer) error
	UploadTar(path string, callbacks UploadCallBacks, reader io.Reader) error
}
