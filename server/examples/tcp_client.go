package main

import (
	"bytes"
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
	"os"
	"path/filepath"

	pb "github.com/Low-Stack-Technologies/temp/server/proto"
	"google.golang.org/protobuf/proto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tcp_client.go <file1> [file2] [file3] ...")
		os.Exit(1)
	}
	
	// Configuration
	serverAddr := "localhost:8080"
	publicKeyURL := "http://localhost:8080/api/public-key"
	ttlSeconds := int32(3600) // 1 hour
	
	// Get public key from server
	fmt.Println("Fetching public key...")
	publicKey, err := getPublicKey(publicKeyURL)
	if err != nil {
		fmt.Printf("Error getting public key: %v\n", err)
		os.Exit(1)
	}
	
	// Generate AES key
	aesKey := make([]byte, 32) // 256 bits
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		fmt.Printf("Error generating AES key: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("Error encrypting AES key: %v\n", err)
		os.Exit(1)
	}
	
	// Connect to server
	fmt.Printf("Connecting to %s...\n", serverAddr)
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Error connecting: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	
	// Send upload request
	uploadReq := &pb.UploadRequest{
		EncryptedAesKey: encryptedAESKey,
		NumFiles:        int32(len(os.Args) - 1),
		TtlSeconds:      ttlSeconds,
	}
	
	if err := sendMessage(conn, uploadReq); err != nil {
		fmt.Printf("Error sending upload request: %v\n", err)
		os.Exit(1)
	}
	
	// Upload each file
	for _, filePath := range os.Args[1:] {
		fmt.Printf("\nUploading %s...\n", filePath)
		
		// Get file info
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Error stating file: %v\n", err)
			continue
		}
		
		// Send file metadata
		metadata := &pb.FileMetadata{
			Filename: filepath.Base(filePath),
			Size:     fileInfo.Size(),
		}
		
		if err := sendMessage(conn, metadata); err != nil {
			fmt.Printf("Error sending metadata: %v\n", err)
			continue
		}
		
		// Read and encrypt file
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			continue
		}
		
		// Send file in chunks
		chunkSize := 64 * 1024 // 64KB chunks
		buffer := make([]byte, chunkSize)
		totalSent := int64(0)
		
		for {
			n, err := file.Read(buffer)
			if n > 0 {
				// Encrypt chunk
				encrypted, err := encryptAES(aesKey, buffer[:n])
				if err != nil {
					fmt.Printf("Error encrypting chunk: %v\n", err)
					break
				}
				
				isLast := err == io.EOF
				chunk := &pb.FileChunk{
					Data:   encrypted,
					IsLast: isLast,
				}
				
				if err := sendMessage(conn, chunk); err != nil {
					fmt.Printf("Error sending chunk: %v\n", err)
					break
				}
				
				totalSent += int64(n)
				fmt.Printf("\rProgress: %d/%d bytes (%.1f%%)",
					totalSent, fileInfo.Size(),
					float64(totalSent)/float64(fileInfo.Size())*100)
				
				if isLast {
					break
				}
			}
			
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("\nError reading file: %v\n", err)
				break
			}
		}
		
		file.Close()
		fmt.Println()
	}
	
	// Receive response
	fmt.Println("\nWaiting for response...")
	response := &pb.UploadResponse{}
	if err := receiveMessage(conn, response); err != nil {
		fmt.Printf("Error receiving response: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("\n✗ Upload failed: %s\n", response.Error)
	}
}

func getPublicKey(url string) (*rsa.PublicKey, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
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
