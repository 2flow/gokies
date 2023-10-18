package filecontainer

import "io"

type IFileManager interface {
	GetFile(path string) (io.ReadCloser, error)
	GetUploadWriter(path string) (io.WriteCloser, error)
	BackupDirectory(path string, writer io.Writer) error
	UploadTar(path string, reader io.Reader) error
}
