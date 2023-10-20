package filecontainer

import (
	compression3 "github.com/2flow/gokies/compression"
	"github.com/2flow/gokies/storageabstraction"
	"github.com/go-kit/log"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

type LockType int
type UploadType int

const (
	LockTypeFileOnly LockType = 0
	LockTypeAll      LockType = 1

	UploadTypeTar UploadType = 0

	LocalTarName = "artifact.tar.gz"
)

type UploadCallBacks struct {
	OnExtractionFinished func(err error)
	OnReadyToExtract     func() error
}

type UploadLockObject struct {
	path            string
	lockType        LockType
	continueObjects []*UploadObject
}

type UploadObject struct {
	localDir       string
	destinationDir string
	uploadType     UploadType
	callbacks      UploadCallBacks
}

type Uploader struct {
	rootDir       string
	uploaderTodos []*UploadLockObject
	logger        log.Logger
	todosLock     sync.Mutex
	fileStorage   storageabstraction.IFileStorage
}

type TarUploader interface {
	io.Writer
	Done()
	Error()
}
type tarUploadWriter struct {
	TarUploader
	uploader     *Uploader
	UploadObject *UploadObject
	file         io.WriteCloser
}

func CreateUploader(rootDir string, logger log.Logger, fileStorage storageabstraction.IFileStorage) *Uploader {
	return &Uploader{logger: logger, rootDir: rootDir, fileStorage: fileStorage}
}

func doesDirectoryExist(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func removeAllContentInDir(dir string) {
	directory, _ := ioutil.ReadDir(dir)
	for _, d := range directory {
		_ = os.RemoveAll(path.Join([]string{dir, d.Name()}...))
	}
}

// provideEmptyDirectory clears or creates the given directory
func provideEmptyDirectory(dir string) {

	if doesDirectoryExist(dir) {
		removeAllContentInDir(dir)
	} else {
		// if directory not exists, create it
		_ = os.MkdirAll(dir, 0777)
	}
}

func (uploader *Uploader) createTempDir() string {
	provideEmptyDirectory(uploader.rootDir)

	dir, err := os.MkdirTemp(uploader.rootDir, "uploaderDir")
	if err != nil {
		uploader.logger.Log("msg", "Unable to create temp dir")
	}

	return dir
}

func (uploader *Uploader) UploadTar(rootPath string, callbacks UploadCallBacks) (TarUploader, error) {
	tempDir := uploader.createTempDir()
	tempPath := path.Join(tempDir, LocalTarName)

	uploadObject := &UploadObject{
		tempDir,
		rootPath,
		UploadTypeTar,
		callbacks,
	}

	file, err := os.Create(tempPath)
	if err != nil {
		return nil, err
	}

	tarWriter := tarUploadWriter{uploader: uploader, UploadObject: uploadObject, file: file}
	return tarWriter, err
}

func (uploader *Uploader) objectUploadFunction(uploadObject *UploadObject, lockObj *UploadLockObject) {

	err := uploadObject.callbacks.OnReadyToExtract()
	if err == nil {
		err = uploader.extractTar(uploadObject)
		uploadObject.callbacks.OnExtractionFinished(err)
	}

	uploader.todosLock.Lock()
	{
		for i, elem := range uploader.uploaderTodos {
			if elem == lockObj {
				todosCount := len(uploader.uploaderTodos)
				uploader.uploaderTodos[i] = uploader.uploaderTodos[todosCount-1]
				uploader.uploaderTodos = uploader.uploaderTodos[:todosCount-1]
				break
			}
		}

		uploader.todosLock.Unlock()
	}

	for _, uploadObject := range lockObj.continueObjects {
		uploader.registerForUpload(uploadObject)
	}

}

func (uploader *Uploader) registerForUpload(uploadObject *UploadObject) {
	uploader.todosLock.Lock()
	defer uploader.todosLock.Unlock()

	for _, todo := range uploader.uploaderTodos {
		if strings.HasPrefix(uploadObject.destinationDir, todo.path) {
			todo.continueObjects = append(todo.continueObjects, uploadObject)
			return
		}
	}

	lockObj := &UploadLockObject{lockType: LockTypeAll, path: uploadObject.destinationDir}
	uploader.uploaderTodos = append(uploader.uploaderTodos, lockObj)

	go uploader.objectUploadFunction(uploadObject, lockObj)
}

func (tarUploadWriter tarUploadWriter) Write(p []byte) (n int, err error) {
	return tarUploadWriter.file.Write(p)
}

/*func (tarUploadWriter tarUploadWriter) Close() error {
	tarUploadWriter.uploader.registerForUpload(tarUploadWriter.UploadObject)
	return tarUploadWriter.file.Close()
}*/

func (tarUploadWriter tarUploadWriter) Done() {
	tarUploadWriter.uploader.registerForUpload(tarUploadWriter.UploadObject)
	_ = tarUploadWriter.file.Close()
}

func (tarUploadWriter tarUploadWriter) Error() {
	_ = tarUploadWriter.file.Close()
}

func (uploader *Uploader) uploadContentFromTar(tarPath string, destinationDir string, tempFile string) ([]string, error) {
	var uploadedFiles []string

	uploader.logger.Log("msg", "Start file extraction ...")

	compression2 := compression3.NewGzipExtractor(uploader.fileStorage)

	/*compression := utils.Compression{
		FolderCallback: func(relativeDir string) {
		},
		FileCallback: func(relativeDir string, fileSize int64, readContent io.Reader) {
			// foreach file in the archive this is called
			tempFile, _ := os.Create(tempFile)
			defer tempFile.Close()

			if written, err := io.Copy(tempFile, readContent); (err != nil) || (written != fileSize) {
				uploader.logger.Log("Unable to copy to temp file")
				return
			}
			if err := tempFile.Sync(); err != nil {
				uploader.logger.Log("Unable to sync temp file")
				return
			}
			if _, err := tempFile.Seek(0, 0); err != nil {
				return
			}

			if err := uploader.fileStorage.Write(destinationDir+"/"+relativeDir, fileSize, tempFile); err == nil {
				uploadedFiles = append(uploadedFiles, relativeDir)
			}
		},
	}*/

	artifactReader, err := os.Open(tarPath)
	if err != nil {
		uploader.logger.Log("msg", "unable to open uploaded artifacts file "+tarPath)
		return uploadedFiles, err
	}
	defer artifactReader.Close()
	/*
		if err := compression.ProcessCompression(artifactReader); err != nil {
			uploader.logger.Log("msg", "unable to process uploaded artifact")
			return uploadedFiles
		}*/
	uploadedFiles, err = compression2.ExtractFromStream(destinationDir, artifactReader)
	if err != nil {
		uploader.logger.Log("msg", "unable to process uploaded artifact: "+err.Error())
	} else {
		uploader.logger.Log("msg", "Finished file extraction")
	}

	return uploadedFiles, err
}

func (uploader *Uploader) removeOldFilesInStorage(uploadObject *UploadObject, uploadedFiles []string) error {
	return uploader.fileStorage.Walk(uploadObject.destinationDir, func(filePath string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			exists := false
			for _, fileInArtifact := range uploadedFiles {
				if fileInArtifact == (filePath) {
					exists = true
					break
				}
			}
			if !exists {
				uploader.fileStorage.DeleteFile(uploadObject.destinationDir + "/" + filePath)
			}
		}
		return nil
	})
}

func (uploader *Uploader) extractTar(uploadObject *UploadObject) error {
	artifactFileName := path.Join(uploadObject.localDir, LocalTarName)
	tempFile := path.Join(uploadObject.localDir, "temp.file")

	uploadedFiles, err := uploader.uploadContentFromTar(artifactFileName, uploadObject.destinationDir, tempFile)
	if err != nil {
		return err
	}

	return uploader.removeOldFilesInStorage(uploadObject, uploadedFiles)
}
