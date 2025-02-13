# Temp - Temporary file storage

TEMP is a simple file storage service that allows you to quickly and easily upload files for a limited time. After the file or files are uploaded you get a simple link to a direct download. The files are automatically deleted after the expiration time.

## Temp CLI

The Temp CLI is the easiest way to use Temp. It allows you to upload files and get a direct download link.

### Installation

You can download the latest version of the CLI from the [releases page](https://github.com/low-stack/temp/releases). Alternatively, you can use the following commands to install the CLI:

### Linux, Darwin (MacOS) & Windows (WSL)

```bash
curl -fsSL https://raw.githubusercontent.com/low-stack-technologies/temp/main/install.sh | bash
```

### Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/low-stack-technologies/temp/main/install.ps1 | iex
```

### Usage

To upload a file, simply run the following command:

```bash
temp file.txt
```

If you are on Windows, you can use the following command:

```powershell
upload-temp file.txt
```

This will upload the file `file.txt` and print the download link. You can also specify an expiration time:

```bash
temp file.txt --expiration 1h
```

This will upload the file `file.txt` and print the download link.

There is also a Windows Explorer Context Menu shortcut available to uploads files with one click.

#### Multiple files

You can upload multiple files by specifying multiple file paths:

```bash
temp file1.txt file2.txt
```

This will upload both `file1.txt` and `file2.txt` and print the download links.

### Use self-hosted Temp Server

If you want to use your own Temp Server, you can set the `TEMP_SERVICE_URL` environment variable to the URL of your Temp Server. For example:

```bash
export TEMP_SERVICE_URL=https://temp.example.com
```

Now you can use the Temp CLI as usual.

## Temp Server

By default, the Temp CLI uses the Temp Server to upload and download files. You can also run your own Temp Server.

### Docker Compose

```yaml
services:
  temp:
    image: ghcr.io/low-stack-technologies/temp:latest
    ports:
      - 8080:8080
    environment:
      - HTTP_PORT=8080 # The port to listen on
      - DATABASE_PATH=/data/temp-server-database.db # The path to the database file
      - BASE_URL=https://temp.example.com # The URL of your Temp Server, no trailing slash

      - STORAGE_PATH=/data # The path to the storage directory
      - MAX_FILE_SIZE=10G # The maximum file size allowed
      - MIN_FREE_SPACE=1G # The minimum free space required to be left after uploading a file

      - DEFAULT_EXPIRATION=24h # Default expiration time
      - MAX_EXPIRATION=72h # Longest allowed expiration time
      - MIN_EXPIRATION=15m # Shortest allowed expiration time
    volumes:
      - ./temp-server-data:/data # The path to the storage directory
    restart: always
```

### Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e HTTP_PORT=8080 \
  -e DATABASE_PATH=/data/temp-server-database.db \
  -e BASE_URL=https://temp.example.com \
  -e STORAGE_PATH=/data \
  -e MAX_FILE_SIZE=10G \
  -e MIN_FREE_SPACE=1G \
  -e DEFAULT_EXPIRATION=24h \
  -e MAX_EXPIRATION=72h \
  -e MIN_EXPIRATION=15m \
  -v ./temp-server-data:/data \
  ghcr.io/low-stack-technologies/temp:latest
```
