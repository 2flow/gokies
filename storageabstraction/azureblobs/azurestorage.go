package filestorage

// https://docs.microsoft.com/en-us/azure/storage/blobs/storage-quickstart-blobs-go?tabs=windows
import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/2flow/gokies/storageabstraction"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	//"github.com/Azure/azure-storage-file-go/azfile"
)

// tAzureFileStorage
//						if the storage URL is "" (Empty string) the default url is used
type tAzureFileStorage struct {
	accountName string
	accountKey  string
	storageURL  string
	credential  *azblob.SharedKeyCredential
	loginCount  int

	loginLock sync.Mutex

	containerName string // will be im-projects
}

type tAzureReadCloser struct {
	readCloser  io.ReadCloser
	fileStorage *tAzureFileStorage
}

func (azureReader *tAzureReadCloser) Read(p []byte) (n int, err error) {
	return azureReader.readCloser.Read(p)
}

func (azureReader *tAzureReadCloser) Close() error {
	azureReader.fileStorage.LogOut()
	return azureReader.readCloser.Close()
}

// File structure
// im-projects
// |-- Project name
// |  |--Version

// NewAzureStorage instantiates a new azur storage connector
func NewAzureStorage(accountName string, accountKey string, containerName string, storageURL string) storageabstraction.IFileStorage {

	if storageURL == "" {
		storageURL = fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
	}

	storage := &tAzureFileStorage{accountName: accountName,
		accountKey:    accountKey,
		storageURL:    storageURL,
		credential:    nil,
		loginCount:    0,
		containerName: containerName}

	return storage
}

// Login to the azure storage system, returns nil if success
func (azureStorage *tAzureFileStorage) LogIn() error {
	azureStorage.loginLock.Lock()
	defer azureStorage.loginLock.Unlock()

	// if currently no task is logged in
	// do the login immediately
	if azureStorage.loginCount == 0 {
		credential, err := azblob.NewSharedKeyCredential(azureStorage.accountName, azureStorage.accountKey)
		if err != nil {
			fmt.Println("Unable to login to azure")
			return err
		}

		azureStorage.credential = credential
	}

	// increase login counter so we know during the logout if something else is using the connection
	azureStorage.loginCount++

	return nil
}

func (azureStorage *tAzureFileStorage) DeleteDirectory(directory string) error {
	azureStorage.LogIn()
	defer azureStorage.LogOut()

	_, containerURL := azureStorage.getContainerURL()

	ctx := context.Background()

	err := azureStorage.Walk(directory, func(path string, info storageabstraction.FileInfo, err error) error {
		if err != nil {
			return err
		}
		blobURL := containerURL.NewBlockBlobURL(blobInfo.Name)
		_, delErr := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})

		return delErr
	})

	return err
}

func (azureStorage *tAzureFileStorage) DeleteFile(fileName string) error {
	azureStorage.LogIn()
	defer azureStorage.LogOut()

	_, blobURL := azureStorage.getBlobURL(fileName)
	ctx := context.Background()
	_, delErr := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionInclude, azblob.BlobAccessConditions{})

	return delErr
}

func (azureStorage *tAzureFileStorage) Walk(directory string, walk storageabstraction.WalkFunc) {
	azureStorage.LogIn()
	defer azureStorage.LogOut()

	_, containerURL := azureStorage.getContainerURL()
	ctx := context.Background()

	var err error = nil

	for marker := (azblob.Marker{}); marker.NotDone(); {
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{Prefix: directory})

		if err != nil {
			log.Printf("Unable to list content: %s\r\n", err.Error())
			err = walk("", storageabstraction.FileInfo{Size: 0, IsDir: false}, err)
		} else {
			// ListBlobs returns the start of the next segment; you MUST use this to get
			// the next segment (after processing the current result segment).
			marker = listBlob.NextMarker

			// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
			for _, blobInfo := range listBlob.Segment.BlobItems {
				err = walk(strings.TrimPrefix(blobInfo.Name, directory), storageabstraction.FileInfo{Size: *blobInfo.Properties.ContentLength, IsDir: false}, nil)

				if err != nil {
					break
				}
			}
		}

		if err != nil {
			break
		}

	}

	return err
}

func (azureStorage *tAzureFileStorage) DownloadFile(fileName string) (io.ReadCloser, error) {
	azureStorage.LogIn()
	defer azureStorage.LogOut()

	_, blobURL := azureStorage.getBlobURL(fileName)

	ctx := context.Background()

	// Here's how to read the blob's data with progress reporting:
	get, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &tAzureReadCloser{get.Body(azblob.RetryReaderOptions{}), azureStorage}, nil
}

func (azureStorage *tAzureFileStorage) getContainerURL() (pipeline.Pipeline, azblob.ContainerURL) {
	p := azblob.NewPipeline(azureStorage.credential, azblob.PipelineOptions{})

	cURL, _ := url.Parse(fmt.Sprintf("%s/%s", azureStorage.storageURL, azureStorage.containerName))

	containerURL := azblob.NewContainerURL(*cURL, p)
	return p, containerURL
}

func (azureStorage *tAzureFileStorage) getBlobURL(fileName string) (pipeline.Pipeline, azblob.BlockBlobURL) {
	p, containerURL := azureStorage.getContainerURL()
	blobURL := containerURL.NewBlockBlobURL(fileName)

	return p, blobURL
}

func (azureStorage *tAzureFileStorage) FileSize(fileName string) (int64, error) {
	azureStorage.LogIn()
	defer azureStorage.LogOut()

	_, blobURL := azureStorage.getBlobURL(fileName)
	ctx := context.Background()

	property, err := blobURL.GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})

	if err != nil {
		fmt.Println("Unable to read blob property")
		return 0, err
	}
	return property.ContentLength(), nil
}

func (azureStorage *tAzureFileStorage) UploadFile(fileName string, fileSize int64, reader io.ReadSeeker) error {
	azureStorage.LogIn()
	defer azureStorage.LogOut()

	// p := azblob.NewPipeline(azureStorage.credential, azblob.PipelineOptions{})

	// // From the Azure portal, get your Storage account blob service URL endpoint.
	// cURL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", azureStorage.accountName, azureStorage.containerName))

	// containerURL := azblob.NewContainerURL(*cURL, p)
	// blobURL := containerURL.NewBlockBlobURL(fileName)

	_, blobURL := azureStorage.getBlobURL(fileName)

	ctx := context.Background()
	// Wrap the request body in a RequestBodyProgress and pass a callback function for progress reporting.
	_, err := blobURL.Upload(ctx, reader,
		azblob.BlobHTTPHeaders{
			ContentType:        "text/html; charset=utf-8",
			ContentDisposition: "attachment",
		}, azblob.Metadata{
			"createdby": "",
		}, azblob.BlobAccessConditions{}, azblob.AccessTierHot, azblob.BlobTagsMap{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (azureStorage *tAzureFileStorage) LogOut() error {
	azureStorage.loginLock.Lock()
	defer azureStorage.loginLock.Unlock()

	azureStorage.loginCount--
	if azureStorage.loginCount == 0 {
		azureStorage.credential = nil
	}

	return nil
}
