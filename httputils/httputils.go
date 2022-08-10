package httputils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// TResponse is the default response to the client
type TResponse struct {
	StatusCode int         `json:"statusCode"`
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	Content    interface{} `json:"content"`
}

// Encode writes the response to the writer and sets the content type
// with status code 200
func (response *TResponse) Encode(w http.ResponseWriter) error {
	encoder := json.NewEncoder(w)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return encoder.Encode(response)
}

// EncodeStatus writes the response to the responseWriter and sets also the status code
func (response *TResponse) EncodeStatus(w http.ResponseWriter, statusCode int) error {
	encoder := json.NewEncoder(w)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return encoder.Encode(response)
}

// GetDefaultResponse creats a successful response with content
func GetDefaultResponse(msg string, content interface{}) *TResponse {
	resp := &TResponse{}
	resp.Status = "Success"
	resp.StatusCode = 0
	resp.Message = msg
	resp.Content = content

	return resp
}

// GetErrorResponse creates a error response with message and errorcode
func GetErrorResponse(msg string, code int, args ...interface{}) *TResponse {
	resp := &TResponse{}
	resp.Status = "Error"
	resp.StatusCode = code
	resp.Message = fmt.Sprintf(msg, args...)
	resp.Content = nil

	return resp
}

func HTTPRoutingErrorHandler(logMsg string, err error) *TResponse {
	log.Printf("%s: %s", logMsg, err.Error())

	return GetErrorResponse(logMsg, 1)
}

// SetContentType Setst the content type to the responsewriter by the filename or the content
func SetContentType(w http.ResponseWriter, file io.Reader, fileName string) {
	SetContentTypeWithName(w, fileName)
}

func SetContentTypeWithName(w http.ResponseWriter, fileName string) bool {

	contentType := ""

	if strings.HasSuffix(fileName, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(fileName, ".js") {
		contentType = "application/javascript"
	} else if (strings.HasSuffix(fileName, ".ico")) || (strings.HasSuffix(fileName, ".cur")) {
		contentType = "image/x-icon"
	} else if strings.HasSuffix(fileName, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(fileName, ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(fileName, ".svg") {
		contentType = "image/svg+xml"
	} else if strings.HasSuffix(fileName, ".json") {
		contentType = "application/json"
	}

	if contentType != "" {
		w.Header().Add("Content-Type", contentType)
		return true
	} else {
		return false
	}
}
