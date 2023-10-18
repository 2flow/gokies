# Gokies

A small private collection of functionality for projects using http, rest and azure as a storage.

## StorageAbstraction

This is a collection of storage targets. All of them implement the IFileStorage interface.
It allows to set a root directory to which all the paths are relative to.

## Compression

Compress and Extract files and folders using gzip and tar.
The compression and extraction is using the storage abstraction.

```go
storage := localstorage.NewLocalStorage("dir")
// Create a new Compression object
compExtractor := compression.NewCompression(storage)
file, err := os.OpenFile("compress.tar.gz", os.O_CREATE|os.O_WRONLY, 0777)
if err != nil {
t.Errorf("Unable to open file: %v", err)
return
}

// compress ./dir/compressDir into compress.tar.gz
err = compressor.CompressDir("compressDir", file

```

## FileContainer (IFileManager & FileManager)

Currently still a mess, trying to organize it better.
Simple task is to have a directory from where files can be served.
With backup and upload functionality:

* ability to upload files to a folder inside this directory using a tar.gz.
* download backup files of a folder inside this directory using a tar.gz.