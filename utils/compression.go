package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
)

// ExtractFileCallback called if the current extraction is a file
type ExtractFileCallback func(relativDir string, fileSize int64, readContent io.Reader)

// ExtractFolderCallback is called if the current extraction is a folder
type ExtractFolderCallback func(relativDir string)

// Compression The type of the compression
//
//	Contains the callbacks
type Compression struct {
	FolderCallback ExtractFolderCallback
	FileCallback   ExtractFileCallback
}

// ProcessCompression Decompress the stream, for each file and folder the corresponding
//
//	Callbacks are called
func (compression *Compression) ProcessCompression(gzipStream io.Reader) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	defer uncompressedStream.Close()

	if err != nil {
		fmt.Println("Unable to get Reader from stream")
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for header, err := tarReader.Next(); err != io.EOF; header, err = tarReader.Next() {
		if err != nil {
			fmt.Println("Extraction failed during Next()")
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			compression.FolderCallback(header.Name)
		case tar.TypeReg:
			compression.FileCallback(header.Name, header.Size, tarReader)
		}
	}

	return nil
}
