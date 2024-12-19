package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"vddk-builder/pkg/builder"
	"vddk-builder/pkg/config"
	"vddk-builder/pkg/k8spermissions"
	"vddk-builder/pkg/registry"
)

var (
	buildLock sync.Mutex // Mutex for controlling access
	isBusy    bool       // Global flag indicating if the server is busy
)

// StartServer initializes and starts the HTTPS server with the provided configuration.
// It sets up the necessary endpoints and handles file uploads and image checks.
//
// Parameters:
//   - cfg: A pointer to the configuration struct containing server settings.
//
// The function performs the following tasks:
//   - Creates the upload directory if it doesn't exist.
//   - Adds an endpoint to check the availability of an image in the registry.
//   - Adds an endpoint to handle file uploads and initiate the build process.
//   - Starts the HTTPS server using the provided certificate and private key.
//
// Endpoints:
//   - /check-image: Checks if an image exists in the registry. Accepts GET requests with an 'image' query parameter.
//   - /upload: Handles file uploads and initiates the build process. Accepts POST requests with a 'file' form field and an optional 'image' query parameter.
//
// The server will respond with appropriate HTTP status codes and messages based on the request and processing results.
func StartServer(cfg *config.Config) {
	// Create upload directory
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Unable to create upload directory: %v", err))
	}

	// Add new endpoint to check image availability
	http.HandleFunc("/check-image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		imageName := r.URL.Query().Get("image")
		if imageName == "" {
			http.Error(w, "Missing 'image' query parameter", http.StatusBadRequest)
			return
		}

		authToken, err := authenticateRequest(cfg, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Check image in the registry
		imageExists, err := registry.CheckImageExists(imageName, cfg.ImageRegistry, authToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error checking image: %v", err), http.StatusInternalServerError)
			return
		}

		if imageExists {
			fmt.Fprintf(w, "Image %s exists in the registry.\n", imageName)
		} else {
			http.Error(w, fmt.Sprintf("Image %s not found in the registry.", imageName), http.StatusNotFound)
		}
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Allow only POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check if the server is busy
		buildLock.Lock()
		if isBusy {
			buildLock.Unlock()
			http.Error(w, "Server is busy processing another build. Please try again later.", http.StatusServiceUnavailable)
			return
		}
		isBusy = true
		buildLock.Unlock()

		// Parse the optional image query parameter
		imageName := r.URL.Query().Get("image")
		if imageName == "" {
			imageName = cfg.ImageName // Use default image name from config
		}

		authToken, err := authenticateRequest(cfg, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			resetBusy()
			return
		}

		// Parse the uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to parse file", http.StatusBadRequest)
			resetBusy()
			return
		}
		defer file.Close()

		// Save the uploaded file
		filePath := filepath.Join(cfg.UploadDir, header.Filename)
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			resetBusy()
			return
		}
		defer dst.Close()

		io.Copy(dst, file)
		fmt.Fprintf(w, "File uploaded successfully: %s\n", filePath)

		// Run the builder in a Goroutine
		go func() {
			builder.BuildAndPushImage(cfg, filePath, imageName, authToken)
			resetBusy()
		}()
	})

	// Start HTTPS server
	fmt.Printf("Starting HTTPS server on port %s\n", cfg.ServerPort)
	err := http.ListenAndServeTLS(":"+cfg.ServerPort, cfg.CAPublicKey, cfg.PrivateKey, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to start HTTPS server: %v", err))
	}
}

func authenticateRequest(cfg *config.Config, r *http.Request) (string, error) {
	if !cfg.RequireAuth {
		return "", nil
	}

	authHeader := r.Header.Get("Authorization")
	authToken := ""
	if strings.HasPrefix(authHeader, "Bearer ") {
		authToken = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if authToken == "" {
		return "", fmt.Errorf("Missing bearer token")
	}

	clientset, err := k8spermissions.CreateClientWithToken(cfg.ImageRegistry, authToken)
	if err != nil {
		return "", fmt.Errorf("Failed to create Kubernetes client")
	}

	allowed, err := k8spermissions.CheckAccessWithToken(clientset, "list", "namespaces")
	if err != nil || !allowed {
		return "", fmt.Errorf("Insufficient permissions to list namespaces")
	}

	return authToken, nil
}

func resetBusy() {
	buildLock.Lock()
	isBusy = false
	buildLock.Unlock()
}
