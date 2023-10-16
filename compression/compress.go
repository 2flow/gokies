package compression

import (
	"archive/tar"
	"compress/gzip"
	"github.com/2flow/gokies/storageabstraction"
	"io"
	"os"
	"path/filepath"
)

type Compressor struct {
	fileStorage storageabstraction.IFileStorage
}

func NewCompression(fileStorage storageabstraction.IFileStorage) *Compressor {
	return &Compressor{
		fileStorage: fileStorage,
	}
}

func (compressor *Compressor) CompressDir(path string, writer io.Writer) error {

	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return compressor.fileStorage.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(filePath)
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := compressor.fileStorage.Read(compressor.fileStorage.Join(path, filePath))
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}
