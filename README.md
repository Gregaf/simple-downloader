# Link Downloader

A concurrent file downloader written in Go. This tool allows you to download multiple files simultaneously from a list of URLs provided in a file.

## Features

- **Concurrent Downloads:** Utilizes Go's concurrency primitives for fast downloads.
- **Easy Configuration:** Simple command-line interface.
- **Makefile Integration:** Streamlined build and test process.

## Requirements

- Go 1.18 or higher

## Getting Started

### Build

To build the application, run:

```bash
make build
```

This will create a `dwnld` binary in the root directory.

### Usage

```bash
./dwnld <destination_directory> <urls_file>
```

- `<destination_directory>`: The directory where downloaded files will be saved.
- `<urls_file>`: A text file containing one URL per line.

Example:
```bash
./dwnld ./downloads urls.txt
```

## Development

### Testing

To run the tests:

```bash
make test
```

### Help

To see available Makefile targets:

```bash
make help
```
