package azureblobs

import (
	"github.com/Azure/azure-storage-blob-go/azblob"
	"io/fs"
	"time"
)

type AzureStorageConfig struct {
	name          string
	key           string
	containerName string
	storageURL    string
}

type AzureFileInfo struct {
	fs.FileInfo
	blobInfo *azblob.BlobItemInternal
}

func (fileInfo *AzureFileInfo) Name() string {
	if fileInfo.blobInfo == nil {
		return ""
	}

	return fileInfo.blobInfo.Name
}

func (fileInfo *AzureFileInfo) Size() int64 {
	if fileInfo.blobInfo == nil {
		return 0
	}
	return *fileInfo.blobInfo.Properties.ContentLength
}

func (fileInfo *AzureFileInfo) Mode() fs.FileMode {
	return fs.ModeIrregular
}

func (fileInfo *AzureFileInfo) ModTime() time.Time {
	if fileInfo.blobInfo == nil {
		return time.Time{}
	}

	return fileInfo.blobInfo.Properties.LastModified
}

func (fileInfo *AzureFileInfo) IsDir() bool {
	return false
}

func (fileInfo *AzureFileInfo) Sys() any {
	return fileInfo.blobInfo
}
