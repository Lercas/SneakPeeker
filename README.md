# SneakPeeker

SneakPeeker is a versatile tool designed to detect and optionally remove suspicious URLs (such as canary tokens) from various file formats. In addition to its original capabilities, it now supports more formats, provides detailed reports, can run in dry-run mode, and leverages parallel processing for faster scanning.

# Key Features
- Multi-format Scanning: Supports scanning of `.pdf`, `.zip`, `.docx`, `.xlsx`, `.pptx`, `.txt` and `.html` files for suspicious URLs.
- Detailed JSON Reports:
- Includes file path, file size, processing time, whether the file is suspicious, and the URLs found.
- Provides a summary with total files scanned, number of suspicious files, and number of normal files.
- URL Removal and Backups:
- Optionally remove suspicious URLs from files.
- Creates a backup of the original file before modifications for safety.
- Dry-Run Mode:
- Analyze files without making any changes, useful for previewing results before final action.
- Parallel Processing:
- Scan multiple files concurrently for faster results.
- Configure the number of worker goroutines to improve performance on large directories.
- Verbose Logging:
- Verbose (`-v`) mode for more detailed output, useful for debugging or tracking the scanning process.
- Custom Ignored Domains:
- Specify domains to ignore during URL scanning via the `--ignore` option.

Installation
1.	Clone the repository:
```bash
git clone https://github.com/Lercas/SneakPeeker.git
cd SneakPeeker
```

2.	Build the tool:
```bash
go build -o sneakpeeker main.go
```

Or install directly:
```bash
go install github.com/Lercas/SneakPeeker@v2.0.0
```

# Usage

```bash
./sneakpeeker [options] FILE_OR_DIRECTORY_PATH
```

## Options

- `-f` – Remove suspicious URLs (e.g., canary tokens) from files.
- `-r report_file` – Specify the name of the JSON report file. Default is report.json.
- `-v` – Enable verbose (debug-level) logging.
- `--dry-run` – Perform the scan without modifying any files.
- `--ignore domains` – Comma-separated list of domains to ignore during scanning.
- `-w workers` – Number of worker goroutines to use for parallel file scanning. Default is 5.

## Examples

Scan a directory and output results:
```bash

```

Scan a file and output results:
```bash

```

Scan a PDF and remove canary tokens:
```bash

```

Scan a file in dry-run mode (no modifications):
```bash

```

Scan with a custom JSON report name:
```bash

```

Scan with a custom JSON report name:
```bash

```

Use multiple workers for faster scanning:
```bash

```

# How It Works

- PDF Files: Extracts and decompresses PDF streams to search for URL patterns.
- ZIP/DOCX/XLSX/PPTX Files: Decompresses and inspects the content files inside for URLs.
- TXT/HTML Files: Directly scans the text content for URLs.

If suspicious URLs are found and the `-f` option is enabled (and not in `--dry-run mode`), they are removed from the file after creating a backup (.bak).

## Example Output

```plaintext
[INFO] The file /path/to/file.docx is suspicious. URLs found:
http://test-example.local
https://another-test-example.local

[DEBUG] The file /path/to/normalfile.txt seems normal.
```

### Generated JSON report example:

```json
{
  "reports": [
    {
      "file_path": "/path/to/file.docx",
      "suspicious": true,
      "found_urls": ["http://test-example.local", "https://another-test-example.local"],
      "file_size": 2048,
      "processed_at": "2024-12-20T15:04:05Z"
    },
    {
      "file_path": "/path/to/normalfile.txt",
      "suspicious": false,
      "found_urls": [],
      "file_size": 512,
      "processed_at": "2024-12-20T15:04:10Z"
    }
  ],
  "summary": {
    "total_files": 2,
    "suspicious_files": 1,
    "normal_files": 1
  }
}
```

# Docker Usage

Build the Docker image:
```bash
docker build -t sneakpeeker:latest .
```

Scan a directory mounted into /data:
```bash
docker run --rm -v /path/to/scan:/data sneakpeeker:latest /data
```

Use -f to remove suspicious URLs:
```bash
docker run --rm -v /path/to/scan:/data sneakpeeker:latest -f /data
```

Generate a JSON report:
```bash
docker run --rm -v /path/to/scan:/data sneakpeeker:latest -r /data/report.json /data
```

Add verbose logging and dry-run mode:
```bash
docker run --rm -v /path/to/scan:/data sneakpeeker:latest -v --dry-run /data
```