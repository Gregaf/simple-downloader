package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gregaf/link-downloader/internal/downloader"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	dir := flag.String("dir", "", "Directory to save downloads")
	file := flag.String("file", "", "File containing URLs to download")
	workers := flag.Int("workers", 5, "Number of concurrent workers")
	retries := flag.Int("retries", 3, "Maximum number of retries per download")
	versionFlag := flag.Bool("version", false, "Print version information")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("Version: %s\nCommit:  %s\nDate:    %s\n", Version, Commit, Date)
		return
	}

	if *dir == "" || *file == "" {
		fmt.Println("Usage: dwnld -dir <directory> -file <urls_file> [-workers <num>] [-retries <num>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg := downloader.Config{
		DownloadDir: *dir,
		URLsFile:    *file,
		MaxWorkers:  *workers,
		MaxRetries:  *retries,
	}

	d := downloader.New(cfg)
	if err := d.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Download complete.")
}
