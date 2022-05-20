package filecontainer

import "io"

type IFileManager interface {
	GetFile(path string) (io.ReadCloser, error)
	UploadTar(path string) (io.WriteCloser, error)
}
