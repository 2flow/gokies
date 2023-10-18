package httputils

import (
	"github.com/2flow/gokies/filecontainer"
	"github.com/2flow/gokies/storageabstraction"
	"io"
	"net/http"
	"strings"
)

type HTTPFileContainer struct {
	FileStorage storageabstraction.IFileStorage
	RootDir     string
}

func (container HTTPFileContainer) ProvideFileHandler() http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		fetchDest := request.Header.Get("Sec-Fetch-Dest") // if index.html --> document otherwise script
		routePath := request.URL.Path

		// if the document is requested return the index.html
		// this should work
		if fetchDest == "document" {
			routePath = "/index.html"
		} else if fetchDest == "" {
			parts := strings.Split(routePath, ".")
			if len(parts) == 1 {
				routePath = "/index.html"
			}
		} else if routePath == "/" {
			routePath = "/index.html"
		}

		reader, err := container.FileStorage.Read(routePath)

		if err != nil {
			HTTPRoutingErrorHandler("Unable to read file", err).EncodeStatus(responseWriter, http.StatusInternalServerError)
			return
		}

		defer reader.Close()
		SetContentType(responseWriter, reader, routePath)
		responseWriter.WriteHeader(http.StatusOK)
		io.Copy(responseWriter, reader)
	})
}

func UploadFileWithMultipart(request *http.Request, fileManager filecontainer.IFileManager, path string) error {
	multipartFileName := "file"
	reader, err := request.MultipartReader()

	if err != nil {
		return err
	}

	for {
		part, err := reader.NextPart()

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if part.FormName() == multipartFileName {
			err = fileManager.UploadTar(path, part)
			_ = part.Close()

			if err != nil {
				return err
			}
		}
	}

	return nil
}
