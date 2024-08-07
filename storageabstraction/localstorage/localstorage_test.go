package localstorage

import (
	"github.com/2flow/gokies/storageabstraction"
	"os"
	"strings"
	"testing"
)

const (
	testTempDir = "testingDir"
)

func TestLocalStorage(t *testing.T) {
	os.RemoveAll(testTempDir)
	defer os.RemoveAll(testTempDir)

	err := createTestDir()
	if err != nil {
		t.Errorf("[TestError] Error creating test dir: %v", err)
		return
	}

	testlocalstorageWrite(t)
	testLocalStorageRead(t)
}

func TestPathJoin(t *testing.T) {
	storage := NewLocalStorage(testTempDir)

	testPathJoin(t, storage, "test/test2", "test", "test2")
	testPathJoin(t, storage, "test/test2/test3", "test", "test2", "test3")
	testPathJoin(t, storage, "test/test2/test3", "test", "test2/test3")

	testPathJoin(t, storage, "test/test2/testA/test3", "test", "test2/testA/", "/test3")

	testPathJoin(t, storage, "test/test2/testA/test3", "test", "test2./testA/", "/test3")
	testPathJoin(t, storage, "test/test2/testA/test3", "test", "test2./testA/", "/test3")
	testPathJoin(t, storage, "test/test2/testA/test3", "./test", "test2./testA/", "/test3")

	testPathJoin(t, storage, "/test/test2", "/test", "test2")

}

func testPathJoin(t *testing.T, storage storageabstraction.IFileStorage, expected string, args ...string) {
	actual := storage.Join(args...)
	if actual != expected {
		t.Error("Join failed", "source: ["+strings.Join(args, ", ")+"]", "Expected:", expected, "Actual:", actual)
	}
}

func testLocalStorageRead(t *testing.T) {
	testContent := "test2"
	storage := NewLocalStorage(testTempDir)

	reader, err := storage.Read("compressDir/test2.txt")
	if err != nil {
		t.Errorf("Error reading test file: %v", err)
		return
	}

	buffer := make([]byte, 1024)
	bytesCount, err := reader.Read(buffer)
	if err != nil {
		t.Errorf("Error reading written test file: %v", err)
		return
	}
	content := string(buffer[:bytesCount])
	if content != testContent {
		t.Errorf("Values are not equal, expected: %v, actual: %v", testContent, content)
		return
	}
}

func testlocalstorageWrite(t *testing.T) {
	storage := NewLocalStorage(testTempDir)

	testText := "Some test Text"
	reader := strings.NewReader(testText)
	err := storage.Write("writeTest1.txt", 4, reader)
	if err != nil {
		t.Errorf("Error writing test file: %v", err)
		return
	}

	// verify written file
	file, err := os.OpenFile(testTempDir+"/writeTest1.txt", os.O_RDONLY, 0777)
	if err != nil {
		t.Errorf("Can not open written test file: %v", err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	bytesCount, err := file.Read(buffer)
	if err != nil {
		t.Errorf("Error reading written test file: %v", err)
		return
	}
	content := string(buffer[:bytesCount])
	if content != testText {
		t.Errorf("Values are not equal, expected: %v, actual: %v", testText, content)
		return
	}
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
