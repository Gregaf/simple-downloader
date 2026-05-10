package downloader

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Config holds the configuration for the downloader.
type Config struct {
	DownloadDir string
	URLsFile    string
	MaxWorkers  int
	MaxRetries  int
	Client      *http.Client
}

// downloadTask represents a single download job.
type downloadTask struct {
	url      string
	filename string
}

// Downloader orchestrates the concurrent download process.
type Downloader struct {
	config Config
	pb     *progressbar.ProgressBar
}

// New creates a new Downloader with the given configuration.
func New(cfg Config) *Downloader {
	if cfg.Client == nil {
		cfg.Client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	return &Downloader{config: cfg}
}

// Run starts the download process.
func (d *Downloader) Run() error {
	if err := os.MkdirAll(d.config.DownloadDir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	urls, err := d.readURLs(d.config.URLsFile)
	if err != nil {
		return fmt.Errorf("error reading URLs file: %v", err)
	}

	if len(urls) == 0 {
		return nil
	}

	d.pb = progressbar.Default(int64(len(urls)), "Downloading files")

	tasks := make(chan downloadTask, len(urls))
	var wg sync.WaitGroup
	
	for i := 0; i < d.config.MaxWorkers; i++ {
		wg.Add(1)
		go d.worker(tasks, &wg)
	}

	for _, u := range urls {
		tasks <- downloadTask{url: u}
	}
	close(tasks)

	wg.Wait()
	fmt.Println() 

	return nil
}

func (d *Downloader) readURLs(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}

	return urls, scanner.Err()
}

func (d *Downloader) worker(tasks <-chan downloadTask, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		filename, _ := url.PathUnescape(filepath.Base(task.url))
		if filename == "" {
			filename = "downloaded_file"
		}
		task.filename = filename

		err := d.downloadWithRetry(task)
		if err != nil {
			fmt.Printf("\nFailed to download %s after %d retries: %v\n", task.url, d.config.MaxRetries, err)
		}
		d.pb.Add(1)
	}
}

func (d *Downloader) downloadWithRetry(task downloadTask) error {
	var lastErr error
	for i := 0; i <= d.config.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * time.Second) 
		}
		lastErr = d.download(task)
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func (d *Downloader) download(task downloadTask) error {
	filePath := filepath.Join(d.config.DownloadDir, task.filename)

	var startOffset int64
	info, err := os.Stat(filePath)
	if err == nil {
		startOffset = info.Size()
	}

	req, err := http.NewRequest("GET", task.url, nil)
	if err != nil {
		return err
	}

	if startOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startOffset))
	}

	resp, err := d.config.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if resp.StatusCode == http.StatusPartialContent {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	out, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
