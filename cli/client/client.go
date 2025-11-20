package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Low-Stack-Technologies/temp/cli/config"
	pb "github.com/Low-Stack-Technologies/temp/server/proto"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/protobuf/proto"
)

// Upload uploads files to the server
func Upload(cfg *config.Config, filePaths []string) error {
	// Parse server URL
	parsedURL, err := url.Parse(cfg.Server)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	// Determine HTTP and TCP ports
	httpPort := 80
	if parsedURL.Scheme == "https" {
		httpPort = 443
	}
	if parsedURL.Port() != "" {
		p, err := strconv.Atoi(parsedURL.Port())
		if err == nil {
			httpPort = p
		}
	}
	tcpPort := httpPort + 1

	// Construct URLs
	publicKeyURL := fmt.Sprintf("%s/api/public-key", cfg.Server)
	tcpHost := parsedURL.Hostname()
	tcpAddr := fmt.Sprintf("%s:%d", tcpHost, tcpPort)

	// Fetch public key
	fmt.Println("Fetching public key...")
	publicKey, err := getPublicKey(publicKeyURL)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Generate AES key
	aesKey := make([]byte, 32) // 256 bits
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Encrypt AES key with RSA
	encryptedAESKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		aesKey,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	// Connect to server
	fmt.Printf("Connecting to %s...\n", tcpAddr)
	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	// Send upload request
	uploadReq := &pb.UploadRequest{
		EncryptedAesKey: encryptedAESKey,
		NumFiles:        int32(len(filePaths)),
		TtlSeconds:      3600, // Default 1 hour
	}

	if err := sendMessage(conn, uploadReq); err != nil {
		return fmt.Errorf("failed to send upload request: %w", err)
	}

	// Upload each file
	for _, filePath := range filePaths {
		if err := uploadFile(conn, filePath, aesKey); err != nil {
			fmt.Printf("\nFailed to upload %s: %v\n", filePath, err)
			continue
		}
	}

	// Receive response
	fmt.Println("\nWaiting for response...")
	response := &pb.UploadResponse{}
	if err := receiveMessage(conn, response); err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	if response.Success {
		fmt.Println("\n✓ Upload successful!")
		for _, file := range response.Files {
			fmt.Printf("\nFile: %s\n", file.Filename)
			fmt.Printf("  ID: %s\n", file.Id)
			fmt.Printf("  Size: %d bytes\n", file.Size)
			fmt.Printf("  Download URL: %s\n", file.DownloadUrl)
			fmt.Printf("  Checksum: %s\n", file.Checksum)
		}
	} else {
		return fmt.Errorf("upload failed: %s", response.Error)
	}

	return nil
}

func uploadFile(conn net.Conn, filePath string, aesKey []byte) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	filename := filepath.Base(filePath)
	fmt.Printf("\nUploading %s...\n", filename)

	// Send file metadata
	metadata := &pb.FileMetadata{
		Filename: filename,
		Size:     fileInfo.Size(),
	}

	if err := sendMessage(conn, metadata); err != nil {
		return err
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create progress bar
	bar := progressbar.DefaultBytes(
		fileInfo.Size(),
		"uploading",
	)

	// Send file in chunks
	chunkSize := 64 * 1024 // 64KB chunks
	buffer := make([]byte, chunkSize)
	var totalSent int64

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			// Encrypt chunk
			encrypted, err := encryptAES(aesKey, buffer[:n])
			if err != nil {
				return fmt.Errorf("failed to encrypt chunk: %w", err)
			}

			totalSent += int64(n)
			isLast := totalSent == fileInfo.Size()
			chunk := &pb.FileChunk{
				Data:   encrypted,
				IsLast: isLast,
			}

			if err := sendMessage(conn, chunk); err != nil {
				return fmt.Errorf("failed to send chunk: %w", err)
			}

			bar.Add(n)

			if isLast {
				break
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}
	
	// Ensure progress bar is finished
	bar.Finish()
	fmt.Println()

	return nil
}

func getPublicKey(url string) (*rsa.PublicKey, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	pemData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

func encryptAES(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func sendMessage(conn net.Conn, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

func receiveMessage(conn net.Conn, msg proto.Message) error {
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return err
	}

	return proto.Unmarshal(data, msg)
}
