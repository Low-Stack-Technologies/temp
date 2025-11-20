package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Low-Stack-Technologies/temp/server/config"
	"github.com/Low-Stack-Technologies/temp/server/crypto"
	"github.com/Low-Stack-Technologies/temp/server/db"
	"github.com/Low-Stack-Technologies/temp/server/handlers"
	"github.com/Low-Stack-Technologies/temp/server/storage"
	"github.com/Low-Stack-Technologies/temp/server/worker"
)

func main() {
	log.Println("Starting file upload server...")
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded: port=%d, storage=%s", cfg.ServerPort, cfg.StorageDirectory)
	
	// Initialize database
	dbPath := filepath.Join(cfg.StorageDirectory, "data.db")
	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	log.Printf("Database initialized: %s", dbPath)
	
	repo := db.NewRepository(database.DB())
	
	// Initialize storage
	stor, err := storage.New(cfg.StorageDirectory)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Printf("Storage initialized: %s", cfg.StorageDirectory)
	
	// Load or generate RSA keys
	keyPair, err := crypto.LoadOrGenerateKeyPair(cfg.RSAKeyPath)
	if err != nil {
		log.Fatalf("Failed to load/generate RSA keys: %v", err)
	}
	log.Printf("RSA keys loaded/generated: %s", cfg.RSAKeyPath)
	
	// Setup HTTP handlers
	mux := http.NewServeMux()
	
	// Upload endpoint
	uploadHandler := handlers.NewHTTPUploadHandler(cfg, repo, stor)
	mux.Handle("/api/upload", uploadHandler)
	
	// Download endpoint
	downloadHandler := handlers.NewFileDownloadHandler(repo, stor)
	mux.HandleFunc("/api/download/", downloadHandler.ServeHTTP)
	
	// Public key endpoint
	publicKeyHandler := handlers.NewPublicKeyHandler(keyPair.Public)
	mux.Handle("/api/public-key", publicKeyHandler)
	
	// Config endpoint
	configHandler := handlers.NewConfigHandler(cfg)
	mux.Handle("/api/config", configHandler)
	
	// File listing endpoint
	listHandler := handlers.NewFileListHandler(cfg, repo)
	mux.Handle("/api/files", listHandler)
	
	// File metadata endpoint
	metadataHandler := handlers.NewFileMetadataHandler(cfg, repo)
	mux.HandleFunc("/api/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/files" || r.URL.Path == "/api/files/" {
			listHandler.ServeHTTP(w, r)
		} else {
			metadataHandler.ServeHTTP(w, r)
		}
	})
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	
	// Serve index.html at root
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./static/index.html")
		} else {
			http.NotFound(w, r)
		}
	})
	
	log.Println("HTTP router configured")
	
	// Setup TCP upload handler
	tcpHandler := handlers.NewTCPUploadHandler(cfg, repo, stor, keyPair.Private)
	log.Println("TCP upload handler configured")
	
	// Start cleanup worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	cleanupWorker := worker.NewCleanupWorker(cfg, repo, stor)
	go cleanupWorker.Start(ctx)
	
	// Start HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	
	// Start TCP listener for custom protocol on a different port
	tcpPort := cfg.ServerPort + 1
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpPort))
		if err != nil {
			log.Printf("Failed to start TCP listener: %v", err)
			return
		}
		defer listener.Close()
		
		log.Printf("TCP upload listener started on port %d", tcpPort)
		
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("TCP accept error: %v", err)
				continue
			}
			
			go func(c net.Conn) {
				if err := tcpHandler.HandleConnection(c); err != nil {
					log.Printf("TCP upload error: %v", err)
				}
			}(conn)
		}
	}()
	
	log.Printf("Server listening on port %d", cfg.ServerPort)
	log.Printf("HTTP endpoints:")
	log.Printf("  POST   %s/api/upload", cfg.BaseURL)
	log.Printf("  GET    %s/api/download/{id}", cfg.BaseURL)
	log.Printf("  GET    %s/api/public-key", cfg.BaseURL)
	log.Printf("  GET    %s/api/files", cfg.BaseURL)
	log.Printf("  GET    %s/api/files/{id}", cfg.BaseURL)
	log.Printf("TCP upload: Connect to port %d", tcpPort)
	
	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		
		log.Println("Shutdown signal received, stopping server...")
		cancel()
		cleanupWorker.Stop()
		httpServer.Shutdown(context.Background())
	}()
	
	// Start HTTP server
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}