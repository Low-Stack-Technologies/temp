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

### Option 1: Run Locally

#### 1. Build the Server

```bash
go build -o file-upload-server .
```

#### 2. Configure Environment

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

#### 3. Run the Server

```bash
./file-upload-server
```

The server will:
- Listen on the configured port for HTTP requests (default: 8080)
- Listen on the next port for TCP uploads (default: 8081)
- Start a background cleanup worker
- Generate RSA keys if they don't exist

> **Note**: The server uses separate ports for HTTP and TCP protocols. HTTP runs on `SERVER_PORT` and TCP runs on `SERVER_PORT + 1`.

### Option 2: Run with Docker

#### Using Docker Compose (Recommended)

```bash
# Build and start the server
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the server
docker-compose down
```

#### Using Docker directly

```bash
# Build the image
docker build -t file-upload-server .

# Run the container
docker run -d \
  -p 8080:8080 \
  -p 8081:8081 \
  -v $(pwd)/storage:/app/storage \
  -e BASE_URL=http://localhost:8080 \
  --name file-upload-server \
  file-upload-server

# View logs
docker logs -f file-upload-server

# Stop the container
docker stop file-upload-server
docker rm file-upload-server
```

**Docker Environment Variables:**
All configuration can be set via environment variables (see `.env.example` for options).

**Volumes:**
- `/app/storage`: Persistent storage for uploaded files and database


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

## Deployment

### Docker Deployment

#### Using Docker Compose (Recommended)

The easiest way to deploy is using Docker Compose:

**Option 1: Download from GitHub Releases**

```bash
# Download the Docker Compose deployment package
wget https://github.com/Low-Stack-Technologies/temp/releases/latest/download/docker-compose-deployment.tar.gz
tar -xzf docker-compose-deployment.tar.gz
cd deployment

# Edit .env.example and save as .env (or use environment variables)
cp .env.example .env
nano .env

# Start the server
docker-compose up -d

# View logs
docker-compose logs -f file-upload-server

# Stop the server
docker-compose down
```

**Option 2: Clone from repository**

```bash
# Clone the repository
git clone https://github.com/Low-Stack-Technologies/temp.git
cd temp/server

# Start the server
docker-compose up -d

# View logs
docker-compose logs -f file-upload-server

# Stop the server
docker-compose down
```

**Customizing Configuration:**

Edit `docker-compose.yml` to change environment variables:

```yaml
environment:
  - SERVER_PORT=8080
  - BASE_URL=https://your-domain.com
  - MAX_UPLOAD_SIZE_MB=2048
  - MAX_FILES_PER_UPLOAD=20
```

Or create a `.env` file:

```bash
SERVER_PORT=8080
BASE_URL=https://your-domain.com
MAX_UPLOAD_SIZE_MB=2048
```

#### Using Docker from GitHub Container Registry

Pre-built images are available on GitHub Container Registry:

```bash
# Pull the latest image
docker pull ghcr.io/low-stack-technologies/temp/server:latest

# Run the container
docker run -d \
  -p 8080:8080 \
  -p 8081:8081 \
  -v $(pwd)/storage:/app/storage \
  -e BASE_URL=http://localhost:8080 \
  --name file-upload-server \
  ghcr.io/low-stack-technologies/temp/server:latest
```

**Available tags:**
- `latest` - Latest stable release
- `v1.0.0` - Specific version
- `1.0` - Major.minor version
- `1` - Major version

#### Docker Compose with Reverse Proxy (Production)

Example with Traefik:

```yaml
version: '3.8'

services:
  file-upload-server:
    image: ghcr.io/low-stack-technologies/temp/server:latest
    restart: unless-stopped
    environment:
      - BASE_URL=https://upload.yourdomain.com
      - MAX_UPLOAD_SIZE_MB=2048
    volumes:
      - ./storage:/app/storage
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.upload.rule=Host(`upload.yourdomain.com`)"
      - "traefik.http.routers.upload.entrypoints=websecure"
      - "traefik.http.routers.upload.tls.certresolver=letsencrypt"
      - "traefik.http.services.upload.loadbalancer.server.port=8080"
    networks:
      - web

networks:
  web:
    external: true
```

### Binary Deployment

Pre-built binaries are available for multiple platforms:

#### Download from GitHub Releases

1. Go to [Releases](https://github.com/Low-Stack-Technologies/temp/releases)
2. Download the binary for your platform:
   - `file-upload-server-linux-amd64.tar.gz` - Linux x86_64
   - `file-upload-server-linux-arm64.tar.gz` - Linux ARM64
   - `file-upload-server-darwin-amd64.tar.gz` - macOS Intel
   - `file-upload-server-darwin-arm64.tar.gz` - macOS Apple Silicon
   - `file-upload-server-windows-amd64.zip` - Windows x86_64

#### Linux/macOS Installation

```bash
# Download and extract (replace with your platform)
wget https://github.com/Low-Stack-Technologies/temp/releases/latest/download/file-upload-server-linux-amd64.tar.gz
tar -xzf file-upload-server-linux-amd64.tar.gz

# Make executable
chmod +x file-upload-server-linux-amd64

# Copy to /usr/local/bin (optional)
sudo mv file-upload-server-linux-amd64 /usr/local/bin/file-upload-server

# Create config
cp .env.example .env
# Edit .env with your settings

# Run
file-upload-server
```

#### Windows Installation

1. Download `file-upload-server-windows-amd64.zip`
2. Extract to a folder (e.g., `C:\file-upload-server`)
3. Copy `.env.example` to `.env` and edit settings
4. Run `file-upload-server-windows-amd64.exe`

#### Systemd Service (Linux)

Create `/etc/systemd/system/file-upload-server.service`:

```ini
[Unit]
Description=File Upload Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/file-upload-server
ExecStart=/usr/local/bin/file-upload-server
Restart=always
RestartSec=5

# Environment
Environment="STORAGE_DIRECTORY=/var/lib/file-upload-server"
Environment="SERVER_PORT=8080"
Environment="BASE_URL=http://localhost:8080"

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable file-upload-server
sudo systemctl start file-upload-server
sudo systemctl status file-upload-server
```

### Kubernetes Deployment

Example Kubernetes manifests:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: file-upload-server
spec:
  replicas: 2
  selector:
    matchLabels:
      app: file-upload-server
  template:
    metadata:
      labels:
        app: file-upload-server
    spec:
      containers:
      - name: server
        image: ghcr.io/low-stack-technologies/temp/server:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8081
          name: tcp
        env:
        - name: BASE_URL
          value: "https://upload.example.com"
        - name: MAX_UPLOAD_SIZE_MB
          value: "2048"
        volumeMounts:
        - name: storage
          mountPath: /app/storage
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: storage
        persistentVolumeClaim:
          claimName: file-upload-storage
---
apiVersion: v1
kind: Service
metadata:
  name: file-upload-server
spec:
  selector:
    app: file-upload-server
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: tcp
    port: 8081
    targetPort: 8081
```

### CI/CD with GitHub Actions

The project includes a GitHub Actions workflow that automatically:

1. **Builds Docker images** for `linux/amd64` and `linux/arm64`
2. **Pushes to GitHub Container Registry** (ghcr.io)
3. **Builds binaries** for:
   - Linux (AMD64, ARM64)
   - macOS (AMD64, ARM64)
   - Windows (AMD64)
4. **Creates GitHub releases** with all artifacts

**Triggering a release:**

```bash
# Tag a new version
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions will automatically build and release
```

**Workflow file:** `.github/workflows/build-release.yml`

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
