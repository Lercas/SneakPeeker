# SneakPeeker

SneakPeeker is a tool designed to detect suspicious URLs in various file formats. It processes ZIP, DOCX, XLSX, PPTX, and PDF files to uncover hidden links that might indicate unauthorized access or data leaks.

## Features

- Scans directories and files for suspicious URLs
- Supports ZIP, DOCX, XLSX, PPTX, and PDF file formats
- Outputs found URLs to the console
- Optionally removes canary tokens from files
- Generates a JSON report file

## Installation

1. Clone the repository:
```bash
git clone https://github.com/Lercas/SneakPeeker.git
cd SneakPeeker
```

2. Build the tool:
```bash
go build -o sneakpeeker main.go
```

## Usage

```bash
./sneakpeeker [-f] [-r report_file] FILE_OR_DIRECTORY_PATH
```

### Parameters

`-f`: (Optional) Remove canary tokens from files.
`-r report_file`: (Optional) Specify the name of the JSON report file. Default is report.json.
`FILE_OR_DIRECTORY_PATH`: Path to the file or directory you want to scan.

## Examples

Scan a directory and output results to the console:

```bash
./sneakpeeker /path/to/directory
```

Scan a file and output results to the console:

```bash
./sneakpeeker /path/to/file.docx
```

Scan a file and remove canary tokens:

```bash
./sneakpeeker -f /path/to/file.pdf
```

Scan a file and generate a JSON report:

```bash
./sneakpeeker -r myreport.json /path/to/file.pdf
```

How It Works

- PDF Files: Scans for URL patterns in decompressed PDF streams.
- ZIP, DOCX, XLSX, PPTX Files: Decompresses the files and scans for URL patterns in the extracted contents.

Example Output

```bash
[INFO] The file /path/to/file.docx is suspicious. URLs found:
http://suspicious-example.local
https://another-suspicious-example.local

[INFO] The file /path/to/anotherfile.pdf seems normal.
```

## Docker Usage

First, build the Docker image:

```bash
docker build -t canarycatcher:1.0 .
```

Then, run the container with volume mounting to access your files. For example, if you want to scan the /path/to/scan directory on your host, you can run:
```bash
docker run --rm -v /path/to/scan:/data canarycatcher:1.0 /data
```

If you want to use additional flags -f to remove canary tokens, you can do so as follows:
```bash
docker run --rm -v /path/to/scan:/data canarycatcher:1.0 -f /data
```

Example command to run with a JSON report:
```bash
docker run --rm -v /path/to/scan:/data canarycatcher:1.0 -r /data/report.json /data
```