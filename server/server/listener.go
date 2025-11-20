package server

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"log"
	"net"
	"net/http"

	"github.com/Low-Stack-Technologies/temp/server/handlers"
)

// UnifiedListener handles both HTTP and TCP protocol connections
type UnifiedListener struct {
	httpHandler *http.ServeMux
	tcpHandler  *handlers.TCPUploadHandler
}

// NewUnifiedListener creates a new unified listener
func NewUnifiedListener(httpHandler *http.ServeMux, tcpHandler *handlers.TCPUploadHandler) *UnifiedListener {
	return &UnifiedListener{
		httpHandler: httpHandler,
		tcpHandler:  tcpHandler,
	}
}

// Serve starts accepting connections
func (l *UnifiedListener) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		
		go l.handleConnection(conn)
	}
}

// handleConnection determines the protocol and routes accordingly
func (l *UnifiedListener) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in connection handler: %v", r)
			conn.Close()
		}
	}()
	
	// Peek at the first few bytes to determine protocol
	reader := bufio.NewReader(conn)
	peek, err := reader.Peek(4)
	if err != nil {
		log.Printf("Error peeking connection: %v", err)
		conn.Close()
		return
	}
	
	// Check if it looks like HTTP (starts with HTTP methods)
	if isHTTP(peek) {
		// Handle as HTTP
		l.handleHTTP(conn, reader)
	} else {
		// Handle as custom TCP protocol
		l.handleTCP(conn, reader)
	}
}

// isHTTP checks if the data looks like an HTTP request
func isHTTP(data []byte) bool {
	// Check for common HTTP methods
	methods := []string{"GET ", "POST", "PUT ", "DELE", "HEAD", "PATC", "OPTI"}
	for _, method := range methods {
		if bytes.HasPrefix(data, []byte(method)) {
			return true
		}
	}
	return false
}

// handleHTTP processes an HTTP connection
func (l *UnifiedListener) handleHTTP(conn net.Conn, reader *bufio.Reader) {
	defer conn.Close()
	
	// Create a connection wrapper that uses our buffered reader
	bufferedConn := &bufferedConn{
		Conn:   conn,
		reader: reader,
	}
	
	// Use http.Server to handle the connection properly
	server := &http.Server{
		Handler: l.httpHandler,
	}
	
	// Serve this single connection
	server.Serve(&singleConnListener{conn: bufferedConn})
}

// handleTCP processes a custom TCP protocol connection
func (l *UnifiedListener) handleTCP(conn net.Conn, reader *bufio.Reader) {
	// Create a connection wrapper
	bufferedConn := &bufferedConn{
		Conn:   conn,
		reader: reader,
	}
	
	// Handle TCP upload
	if err := l.tcpHandler.HandleConnection(bufferedConn); err != nil {
		log.Printf("TCP upload error: %v", err)
	}
}

// bufferedConn wraps a connection with a buffered reader
type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

// Read reads from the buffered reader
func (bc *bufferedConn) Read(p []byte) (int, error) {
	return bc.reader.Read(p)
}

// singleConnListener implements net.Listener for a single connection
type singleConnListener struct {
	conn net.Conn
	used bool
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	if l.used {
		// Return a closed connection to signal EOF
		return nil, net.ErrClosed
	}
	l.used = true
	return l.conn, nil
}

func (l *singleConnListener) Close() error {
	return nil
}

func (l *singleConnListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}

// responseWriter is a minimal http.ResponseWriter implementation
type responseWriter struct {
	conn   net.Conn
	header http.Header
}

func (rw responseWriter) Header() http.Header {
	if rw.header == nil {
		rw.header = make(http.Header)
	}
	return rw.header
}

func (rw responseWriter) Write(data []byte) (int, error) {
	return rw.conn.Write(data)
}

func (rw responseWriter) WriteHeader(statusCode int) {
	// Not implemented for simplicity
}

// CreateKeyPair generates or loads RSA keys
func CreateKeyPair(keyPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// This is a placeholder - actual implementation is in crypto package
	return nil, nil, nil
}
