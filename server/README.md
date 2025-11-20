# File Upload Server

A high-performance dual-protocol file upload server supporting both HTTP and encrypted TCP uploads with automatic file expiration.

## Features

- **Dual Protocol Support**: Upload files via HTTP or direct TCP connection
- **Public Key Encryption**: TCP uploads use RSA-2048 + AES-256-GCM hybrid encryption
- **Multi-File Uploads**: Upload multiple files in a single request
- **Automatic Expiration**: Files are automatically deleted after a specified TTL
- **Download URLs**: Get instant download links for uploaded files
- **SQLite Database**: Lightweight metadata storage
- **Security**: Protection against path traversal, SQL injection, and other attacks
- **OpenAPI Documentation**: Full API specification included

## Quick Start

### 1. Build the Server

```bash
go build -o file-upload-server .
```

### 2. Configure Environment

Copy `.env.example` to `.env` and adjust settings:

```bash
cp .env.example .env
```

Key configuration options:
- `STORAGE_DIRECTORY`: Where files and database are stored (default: `./storage`)
- `SERVER_PORT`: Server port (default: `8080`)
- `BASE_URL`: Base URL for download links (default: `http://localhost:8080`)
- `MAX_UPLOAD_SIZE_MB`: Maximum file size in MB (default: `1024`)
- `MAX_FILES_PER_UPLOAD`: Maximum files per upload (default: `10`)

### 3. Run the Server

```bash
./file-upload-server
```

The server will:
- Listen on the configured port for HTTP requests (default: 8080)
- Listen on the next port for TCP uploads (default: 8081)
- Start a background cleanup worker
- Generate RSA keys if they don't exist

> **Note**: The server uses separate ports for HTTP and TCP protocols. HTTP runs on `SERVER_PORT` and TCP runs on `SERVER_PORT + 1`.

## HTTP API

### Upload Files

**Endpoint**: `POST /api/upload`

**Content-Type**: `multipart/form-data`

**Parameters**:
- `file`: One or more files to upload (max 10)
- `ttl_seconds`: Time-to-live in seconds (60 to 2,592,000)

**Example**:
```bash
# Upload single file
curl -X POST http://localhost:8080/api/upload \
  -F "file=@document.pdf" \
  -F "ttl_seconds=3600"

# Upload multiple files
curl -X POST http://localhost:8080/api/upload \
  -F "file=@file1.txt" \
  -F "file=@file2.jpg" \
  -F "file=@file3.pdf" \
  -F "ttl_seconds=7200"
```

**Response**:
```json
{
  "success": true,
  "files": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "filename": "document.pdf",
      "size": 1048576,
      "mime_type": "application/pdf",
      "download_url": "http://localhost:8080/api/download/550e8400-e29b-41d4-a716-446655440000",
      "uploaded_at": "2025-11-20T18:00:00Z",
      "expires_at": "2025-11-20T19:00:00Z",
      "checksum": "a1b2c3d4..."
    }
  ],
  "expires_at": "2025-11-20T19:00:00Z"
}
```

### Download File

**Endpoint**: `GET /api/download/{id}`

**Example**:
```bash
curl http://localhost:8080/api/download/550e8400-e29b-41d4-a716-446655440000 \
  -o downloaded-file.pdf
```

### Get Public Key (for TCP uploads)

**Endpoint**: `GET /api/public-key`

**Example**:
```bash
curl http://localhost:8080/api/public-key
```

### List Files

**Endpoint**: `GET /api/files?limit=50&offset=0`

**Example**:
```bash
curl http://localhost:8080/api/files
```

### Get File Metadata

**Endpoint**: `GET /api/files/{id}`

**Example**:
```bash
curl http://localhost:8080/api/files/550e8400-e29b-41d4-a716-446655440000
```

## TCP Protocol

The TCP protocol uses Protocol Buffers for message serialization and hybrid encryption (RSA + AES) for security.

> **Note**: TCP uploads connect to `SERVER_PORT + 1` (e.g., if HTTP is on 8080, TCP is on 8081)

### Protocol Flow

1. **Get Public Key**: Retrieve server's RSA public key via HTTP
2. **Generate AES Key**: Client generates a random 256-bit AES key
3. **Encrypt AES Key**: Encrypt the AES key with server's RSA public key
4. **Send Upload Request**: Send protobuf message with encrypted AES key, file count, and TTL
5. **For Each File**:
   - Send file metadata (filename, size)
   - Encrypt file data with AES key
   - Send encrypted chunks
6. **Receive Response**: Get file IDs and download URLs

### Message Format

All messages are length-prefixed:
```
[4 bytes: message length (big-endian uint32)][N bytes: protobuf message]
```

See `proto/upload.proto` for full message definitions.

### Example (Conceptual)

```go
// 1. Get public key
publicKey := getPublicKey("http://localhost:8080/api/public-key")

// 2. Generate and encrypt AES key
aesKey := generateAESKey()
encryptedAESKey := rsaEncrypt(publicKey, aesKey)

// 3. Connect and send upload request
conn := dial("localhost:8080")
sendMessage(conn, &UploadRequest{
    EncryptedAesKey: encryptedAESKey,
    NumFiles: 2,
    TtlSeconds: 3600,
})

// 4. For each file
for _, file := range files {
    sendMessage(conn, &FileMetadata{
        Filename: file.name,
        Size: file.size,
    })
    
    // Send encrypted chunks
    for chunk := range file.chunks {
        encrypted := aesEncrypt(aesKey, chunk)
        sendMessage(conn, &FileChunk{
            Data: encrypted,
            IsLast: chunk.isLast,
        })
    }
}

// 5. Receive response
response := receiveMessage(conn)
```

For a complete working example, see `examples/tcp_client.go`.

## Security Features

### Input Validation
- **Filename Sanitization**: Removes path traversal characters (`..`, `/`, `\`, null bytes)
- **File Size Limits**: Configurable maximum file size per upload
- **TTL Validation**: Enforces minimum and maximum TTL values
- **File Count Limits**: Maximum number of files per upload

### SQL Injection Protection
- All database queries use prepared statements
- No string concatenation in SQL queries

### Encryption
- **RSA-2048**: For encrypting AES keys in TCP uploads
- **AES-256-GCM**: For encrypting file data in TCP uploads
- **SHA-256**: For file integrity checksums

### Automatic Cleanup
- Background worker periodically scans for expired files
- Deletes both file data and database records
- Configurable cleanup interval

## Architecture

```
┌─────────────────────────────────────────┐
│         Unified TCP Listener            │
│         (Port 8080)                     │
└────────────┬────────────────────────────┘
             │
             ├─ Protocol Detection
             │
      ┌──────┴──────┐
      │             │
┌─────▼─────┐ ┌────▼────────┐
│   HTTP    │ │     TCP     │
│  Handler  │ │   Handler   │
└─────┬─────┘ └────┬────────┘
      │            │
      │            ├─ Decrypt (RSA+AES)
      │            │
      └────────┬───┘
               │
        ┌──────▼───────┐
        │   Storage    │
        │   Service    │
        └──────┬───────┘
               │
        ┌──────▼───────┐
        │   SQLite DB  │
        └──────────────┘

┌─────────────────────┐
│  Cleanup Worker     │
│  (Background)       │
└─────────────────────┘
```

## Project Structure

```
server/
├── main.go                 # Entry point
├── config/                 # Configuration management
│   └── config.go
├── db/                     # Database layer
│   ├── database.go
│   ├── repository.go
│   └── schema.sql
├── crypto/                 # Encryption utilities
│   ├── keys.go
│   └── encryption.go
├── storage/                # File storage
│   └── storage.go
├── handlers/               # Request handlers
│   ├── http_upload.go
│   ├── tcp_upload.go
│   ├── file_download.go
│   ├── public_key.go
│   └── file_list.go
├── server/                 # Server infrastructure
│   ├── listener.go
│   └── http_router.go
├── worker/                 # Background workers
│   └── cleanup.go
├── proto/                  # Protocol Buffers
│   ├── upload.proto
│   └── upload.pb.go
└── api/                    # API documentation
    └── openapi.yaml
```

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_DIRECTORY` | `./storage` | Base directory for uploads and database |
| `SERVER_PORT` | `8080` | Port for TCP listener |
| `BASE_URL` | `http://localhost:8080` | Base URL for download links |
| `MAX_UPLOAD_SIZE_MB` | `1024` | Maximum file size in MB |
| `MAX_FILES_PER_UPLOAD` | `10` | Maximum files per upload |
| `RSA_KEY_PATH` | `./storage/private_key.pem` | Path to RSA private key |
| `CLEANUP_INTERVAL_SECONDS` | `60` | Cleanup worker interval |
| `MIN_TTL_SECONDS` | `60` | Minimum TTL (1 minute) |
| `MAX_TTL_SECONDS` | `2592000` | Maximum TTL (30 days) |

## Development

### Prerequisites
- Go 1.25.4 or later
- Protocol Buffers compiler (`protoc`)
- `protoc-gen-go` plugin

### Building
```bash
go build -o file-upload-server .
```

### Running Tests
```bash
go test ./...
```

### Generating Protobuf Code
```bash
protoc --go_out=. --go_opt=paths=source_relative proto/upload.proto
```

## Troubleshooting

### Server won't start
- Check if port 8080 is already in use
- Verify `STORAGE_DIRECTORY` is writable
- Check logs for specific error messages

### Files not being cleaned up
- Verify `CLEANUP_INTERVAL_SECONDS` is set correctly
- Check server logs for cleanup worker messages
- Ensure system time is correct

### TCP uploads failing
- Verify RSA keys are generated correctly
- Check that AES key encryption is using the correct public key
- Ensure protobuf messages are properly formatted

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
