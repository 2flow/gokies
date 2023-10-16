package compression

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/2flow/gokies/storageabstraction"
	"io"
)

// ExtractFileCallback called if the current extraction is a file
type ExtractFileCallback func(relativeDir string, fileSize int64, readContent io.Reader)

// ExtractFolderCallback is called if the current extraction is a folder
type ExtractFolderCallback func(relativeDir string)

// GzipExtractor The type of the compression
//
//	Contains the callbacks
type GzipExtractor struct {
	storage storageabstraction.IFileStorage
}

// NewGzipExtractor Creates a new GzipExtractor object
func NewGzipExtractor(storage storageabstraction.IFileStorage) *GzipExtractor {
	return &GzipExtractor{
		storage: storage,
	}
}

// ProcessCompression Decompress the stream, for each file and folder the corresponding
//
//	Callbacks are called
func (extractor *GzipExtractor) ProcessCompression(directory string, gzipStream io.Reader) error {
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
		case tar.TypeReg:
			path := extractor.storage.Join(directory, header.Name)
			err := extractor.storage.Write(path, header.Size, &readerSeeker{reader: tarReader})
			if err != nil {
				return err
			}
		case tar.TypeDir:
		}
	}

	return nil
}

type readerSeeker struct {
	io.ReadSeeker
	reader io.Reader
}

func (reader *readerSeeker) Read(p []byte) (n int, err error) {
	return reader.reader.Read(p)
}

func (reader *readerSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("not implemented")
}
