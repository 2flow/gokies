package compression

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/2flow/gokies/storageabstraction"
	"io"
	"os"
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

// ExtractFromStream Decompress the stream, for each file and folder the corresponding
//
//	Callbacks are called
func (extractor *GzipExtractor) ExtractFromStream(directory string, gzipStream io.Reader) ([]string, error) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	defer uncompressedStream.Close()

	var extractedFiles []string

	if err != nil {
		fmt.Println("Unable to get Reader from stream")
		return extractedFiles, err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for header, err := tarReader.Next(); err != io.EOF; header, err = tarReader.Next() {
		if err != nil {
			fmt.Println("Extraction failed during Next()")
			return extractedFiles, err
		}
		switch header.Typeflag {
		case tar.TypeReg:
			path := extractor.storage.Join(directory, header.Name)
			extractedFiles = append(extractedFiles, header.Name)

			tempReader, err := newTempReaderSeeker(header.Size, tarReader)
			if err != nil {
				return extractedFiles, err
			}
			err = extractor.storage.Write(path, header.Size, tempReader)
			if err != nil {
				_ = tempReader.Close()
				return extractedFiles, err
			}
			_ = tempReader.Close()

		case tar.TypeDir:
		}
	}

	return extractedFiles, nil
}

type tempReaderSeeker struct {
	io.ReadSeekCloser
	file *os.File
}

func newTempReaderSeeker(fileSize int64, reader io.Reader) (*tempReaderSeeker, error) {
	file, err := os.CreateTemp("", "tempUploader")
	if err != nil {
		return nil, err
	}

	if written, err := io.Copy(file, reader); (err != nil) || (written != fileSize) {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("unable to copy to temp file")
	}

	if err := file.Sync(); err != nil {
		return nil, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	return &tempReaderSeeker{
		file: file,
	}, nil
}

func (reader *tempReaderSeeker) Read(p []byte) (n int, err error) {
	return reader.file.Read(p)
}

func (reader *tempReaderSeeker) Seek(offset int64, whence int) (int64, error) {
	return reader.file.Seek(offset, whence)
}

func (reader *tempReaderSeeker) Close() error {
	err := reader.file.Close()
	if err != nil {
		return err
	}

	err = os.Remove(reader.file.Name())
	return err
}
