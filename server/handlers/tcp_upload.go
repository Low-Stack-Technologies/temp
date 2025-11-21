package handlers

import (
	"bytes"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Low-Stack-Technologies/temp/server/config"
	"github.com/Low-Stack-Technologies/temp/server/crypto"
	"github.com/Low-Stack-Technologies/temp/server/db"
	pb "github.com/Low-Stack-Technologies/temp/server/proto"
	"github.com/Low-Stack-Technologies/temp/server/storage"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// TCPUploadHandler handles TCP file uploads with encryption
type TCPUploadHandler struct {
	cfg        *config.Config
	repo       *db.Repository
	storage    *storage.Storage
	privateKey *rsa.PrivateKey
}

// NewTCPUploadHandler creates a new TCP upload handler
func NewTCPUploadHandler(cfg *config.Config, repo *db.Repository, stor *storage.Storage, privateKey *rsa.PrivateKey) *TCPUploadHandler {
	return &TCPUploadHandler{
		cfg:        cfg,
		repo:       repo,
		storage:    stor,
		privateKey: privateKey,
	}
}

// HandleConnection handles a TCP upload connection
func (h *TCPUploadHandler) HandleConnection(conn net.Conn) error {
	defer conn.Close()
	
	// Read upload request
	uploadReq, err := h.readMessage(conn, &pb.UploadRequest{})
	if err != nil {
		return h.sendError(conn, "failed to read upload request: "+err.Error())
	}
	
	req := uploadReq.(*pb.UploadRequest)
	
	// Validate TTL
	if req.TtlSeconds < int32(h.cfg.MinTTLSeconds) || req.TtlSeconds > int32(h.cfg.MaxTTLSeconds) {
		return h.sendError(conn, fmt.Sprintf("ttl_seconds must be between %d and %d", h.cfg.MinTTLSeconds, h.cfg.MaxTTLSeconds))
	}
	
	// Validate file count
	if req.NumFiles <= 0 || req.NumFiles > int32(h.cfg.MaxFilesPerUpload) {
		return h.sendError(conn, fmt.Sprintf("num_files must be between 1 and %d", h.cfg.MaxFilesPerUpload))
	}
	
	// Decrypt AES key
	aesKey, err := crypto.DecryptRSA(h.privateKey, req.EncryptedAesKey)
	if err != nil {
		return h.sendError(conn, "failed to decrypt AES key: "+err.Error())
	}
	
	uploadedAt := time.Now()
	expiresAt := uploadedAt.Add(time.Duration(req.TtlSeconds) * time.Second)
	
	var uploadedFiles []*pb.FileInfo
	
	// Process each file
	for i := int32(0); i < req.NumFiles; i++ {
		// Read file metadata
		metadataMsg, err := h.readMessage(conn, &pb.FileMetadata{})
		if err != nil {
			return h.sendError(conn, "failed to read file metadata: "+err.Error())
		}
		metadata := metadataMsg.(*pb.FileMetadata)
		
		// Validate file size
		if metadata.Size > h.cfg.MaxUploadSizeBytes() {
			return h.sendError(conn, fmt.Sprintf("file %s exceeds maximum size", metadata.Filename))
		}
		
		// Read and decrypt file chunks
		var fileData bytes.Buffer
		var bytesRead int64
		
		for {
			chunkMsg, err := h.readMessage(conn, &pb.FileChunk{})
			if err != nil {
				return h.sendError(conn, "failed to read file chunk: "+err.Error())
			}
			chunk := chunkMsg.(*pb.FileChunk)
			
			// Decrypt chunk
			decrypted, err := crypto.DecryptAES(aesKey, chunk.Data)
			if err != nil {
				return h.sendError(conn, "failed to decrypt chunk: "+err.Error())
			}
			
			fileData.Write(decrypted)
			bytesRead += int64(len(decrypted))
			
			if chunk.IsLast {
				break
			}
		}
		
		// Save file
		storagePath, checksum, size, err := h.storage.SaveFile(&fileData, metadata.Filename)
		if err != nil {
			return h.sendError(conn, "failed to save file: "+err.Error())
		}
		
		// Create database record
		fileID := uuid.New().String()
		record := &db.FileRecord{
			ID:           fileID,
			Filename:     metadata.Filename,
			Size:         size,
			MimeType:     "application/octet-stream", // TCP doesn't provide MIME type
			StoragePath:  storagePath,
			UploadMethod: "tcp",
			UploadedAt:   uploadedAt,
			TTLSeconds:   int(req.TtlSeconds),
			ExpiresAt:    expiresAt,
			Checksum:     checksum,
		}
		
		if err := h.repo.CreateFileRecord(record); err != nil {
			h.storage.DeleteFile(storagePath)
			return h.sendError(conn, "failed to save file metadata: "+err.Error())
		}
		
		// Add to response
		downloadPath := fmt.Sprintf("/api/download/%s", fileID)
		uploadedFiles = append(uploadedFiles, &pb.FileInfo{
			Id:          fileID,
			Filename:    metadata.Filename,
			Size:        size,
			DownloadUrl: downloadPath,
			Checksum:    checksum,
		})
	}
	
	// Send success response
	response := &pb.UploadResponse{
		Success:   true,
		Files:     uploadedFiles,
		ExpiresAt: expiresAt.Unix(),
	}
	
	return h.sendMessage(conn, response)
}

// readMessage reads a length-prefixed protobuf message
func (h *TCPUploadHandler) readMessage(conn net.Conn, msg proto.Message) (proto.Message, error) {
	// Read message length (4 bytes)
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	
	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}
	
	// Unmarshal protobuf
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	
	return msg, nil
}

// sendMessage sends a length-prefixed protobuf message
func (h *TCPUploadHandler) sendMessage(conn net.Conn, msg proto.Message) error {
	// Marshal protobuf
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	
	// Write length prefix
	length := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}
	
	// Write message data
	if _, err := conn.Write(data); err != nil {
		return err
	}
	
	return nil
}

// sendError sends an error response
func (h *TCPUploadHandler) sendError(conn net.Conn, message string) error {
	response := &pb.UploadResponse{
		Success: false,
		Error:   message,
	}
	h.sendMessage(conn, response)
	return fmt.Errorf(message)
}
