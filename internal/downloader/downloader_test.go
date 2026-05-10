package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloader_Run(t *testing.T) {
	// Setup a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		content := "Hello, World! This is a test file content."

		if rangeHeader != "" {
			var start int
			_, err := fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if start >= len(content) {
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}
			w.WriteHeader(http.StatusPartialContent)
			fmt.Fprint(w, content[start:])
		} else {
			fmt.Fprint(w, content)
		}
	}))
	defer server.Close()

	// Create a temporary directory for downloads
	tmpDir, err := os.MkdirTemp("", "downloader_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary URLs file
	urlsFile := filepath.Join(tmpDir, "urls.txt")
	err = os.WriteFile(urlsFile, []byte(server.URL+"/file1.txt\n"+server.URL+"/file2.txt"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		DownloadDir: tmpDir,
		URLsFile:    urlsFile,
		MaxWorkers:  2,
		MaxRetries:  1,
	}

	d := New(cfg)
	err = d.Run()
	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}

	// Verify downloads
	files := []string{"file1.txt", "file2.txt"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", f)
		}
		content, _ := os.ReadFile(path)
		if string(content) != "Hello, World! This is a test file content." {
			t.Errorf("File %s content mismatch", f)
		}
	}
}

func TestDownloader_Resume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := "Full content of the file."
		rangeHeader := r.Header.Get("Range")

		if rangeHeader != "" {
			var start int
			fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
			w.WriteHeader(http.StatusPartialContent)
			fmt.Fprint(w, content[start:])
		} else {
			fmt.Fprint(w, content)
		}
	}))
	defer server.Close()

	tmpDir, _ := os.MkdirTemp("", "resume_test")
	defer os.RemoveAll(tmpDir)

	// Create a partial file
	filePath := filepath.Join(tmpDir, "resume.txt")
	os.WriteFile(filePath, []byte("Full content"), 0644) // Missing " of the file."

	urlsFile := filepath.Join(tmpDir, "urls.txt")
	os.WriteFile(urlsFile, []byte(server.URL+"/resume.txt"), 0644)

	cfg := Config{
		DownloadDir: tmpDir,
		URLsFile:    urlsFile,
		MaxWorkers:  1,
	}

	d := New(cfg)
	d.Run()

	content, _ := os.ReadFile(filePath)
	expected := "Full content of the file."
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestReadURLs(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "urls_test")
	defer os.Remove(tmpFile.Name())

	content := "http://example.com/1\n\n# comment\n  http://example.com/2  \n"
	tmpFile.WriteString(content)
	tmpFile.Close()

	d := New(Config{})
	urls, err := d.readURLs(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(urls) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(urls))
	}
	if urls[0] != "http://example.com/1" || urls[1] != "http://example.com/2" {
		t.Errorf("URL mismatch")
	}
}
