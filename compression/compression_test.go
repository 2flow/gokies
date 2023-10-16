package compression

import (
	"github.com/2flow/gokies/storageabstraction/localstorage"
	"os"
	"testing"
)

const (
	testTempDir = "testingDir"
)

func TestCompression(t *testing.T) {
	os.RemoveAll(testTempDir)
	defer os.RemoveAll(testTempDir)

	err := createTestDir()
	if err != nil {
		t.Errorf("[TestError] Error creating test dir: %v", err)
		return
	}

	storage := localstorage.NewLocalStorage(testTempDir)
	compressor := NewCompression(storage)

	file, err := os.OpenFile(testTempDir+"/compressDir.tar.gz", os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		t.Errorf("[TestError] Error creating test compress file: %v", err)
		return
	}

	err = compressor.CompressDir("compressDir", file)
	if err != nil {
		t.Errorf("Error compressing dir: %v", err)
		return
	}

	err = os.RemoveAll(testTempDir + "/compressDir")
	if err != nil {
		t.Errorf("[TestError] Error removing test dir: %v", err)
		return
	}
	file.Close()

	file, err = os.OpenFile(testTempDir+"/compressDir.tar.gz", os.O_RDONLY, 0777)
	if err != nil {
		t.Errorf("[TestError] Error opening test compress file: %v", err)
		return
	}

	os.RemoveAll(testTempDir + "/compressDir")
	extractor := NewGzipExtractor(storage)

	err = extractor.ProcessCompression("extractDir", file)
	if err != nil {
		t.Errorf("Error extracting file: %v", err)
		return
	}

	fileName := testTempDir + "/extractDir/test.txt"
	_, err = os.ReadFile(fileName)
	if err != nil {
		t.Errorf("Error reading extracted file (%s): %v", fileName, err)
	}

	fileName = testTempDir + "/extractDir/subDir/test3.txt"
	_, err = os.ReadFile(fileName)
	if err != nil {
		t.Errorf("Error reading extracted file (%s): %v", fileName, err)
	}

	file.Close()
}

func createTestDir() error {
	err := os.MkdirAll(testTempDir+"/compressDir", 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(testTempDir+"/compressDir/test.txt", []byte("test"), 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(testTempDir+"/compressDir/test2.txt", []byte("test2"), 0777)
	if err != nil {
		return err
	}

	// create sub dir
	err = os.MkdirAll(testTempDir+"/compressDir/subDir", 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(testTempDir+"/compressDir/subDir/test3.txt", []byte("test3"), 0777)
	if err != nil {
		return err
	}

	return nil
}
