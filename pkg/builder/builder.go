package builder

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"vddk-builder/pkg/config"
)

const dirPerm = 0755

// BuildAndPushImage builds a Docker image from a tar.gz file and pushes it to a Docker registry.
// It performs the following steps:
// 1. Creates a temporary directory for extraction.
// 2. Extracts the contents of the tar.gz file to the temporary directory.
// 3. Builds a Docker image from the extracted contents.
// 4. Pushes the Docker image to the specified registry.
// 5. Cleans up the temporary directory and the tar.gz file.
//
// Parameters:
// - cfg: Configuration object containing image registry and default image name.
// - filePath: Path to the tar.gz file to be extracted and used for building the image.
// - imageName: Name of the Docker image to be built. If empty, the default name from the configuration is used.
// - authToken: The authentication token for the registry.
func BuildAndPushImage(cfg *config.Config, filePath, imageName, authToken string) {
	tmpDir := filepath.Join(".", "tmp")
	if err := os.MkdirAll(tmpDir, dirPerm); err != nil {
		log.Printf("Failed to create temporary directory: %v\n", err)
		return
	}

	// Define the extracted directory under tmp
	extractedDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractedDir, dirPerm); err != nil {
		log.Printf("Failed to create extraction directory: %v\n", err)
		return
	}

	// Defer cleanup for extractedDir and tar.gz file
	defer func() {
		log.Println("Cleaning up...")
		if err := os.RemoveAll(extractedDir); err != nil {
			log.Printf("Failed to remove extracted directory: %v\n", err)
		}
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to remove tar.gz file: %v\n", err)
		}
	}()

	// Extract the tar.gz file
	log.Println("Extracting uploaded file...")
	if err := extractTarGz(filePath, extractedDir); err != nil {
		log.Printf("Failed to extract archive: %v\n", err)
		return
	}

	// Set image name and tag
	if imageName == "" {
		imageName = cfg.ImageName
	}
	imageTag := fmt.Sprintf("%s/%s", cfg.ImageRegistry, imageName)

	// Build the image
	if err := buildImage(imageTag, extractedDir); err != nil {
		log.Printf("Failed to build image: %v\n", err)
		return
	}

	// Push the image to the registry
	if err := pushImage(imageTag, authToken); err != nil {
		log.Printf("Failed to push image: %v\n", err)
		return
	}

	log.Println("Image build and push completed successfully.")
}

// extractTarGz extracts a .tar.gz file to a destination directory
func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %v", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		hdr, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading tar.gz file: %v", err)
		}

		target := filepath.Join(dest, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %v", err)
			}
			outFile.Close()
		}
	}

	return nil
}

// buildImage is an internal method to build the image using podman
func buildImage(imageTag, contextDir string) error {
	cmd := exec.Command("podman", "build", "-f", "Containerfile.vddk", "-t", imageTag, contextDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build image: %w\n%s", err, output)
	}
	return nil
}

// pushImage is an internal method to push the image to the registry
func pushImage(imageTag, authToken string) error {
	// Construct the skopeo command
	args := []string{"copy", "--dest-tls-verify=false"}
	if authToken != "" {
		args = append(args, "--dest-creds", fmt.Sprintf(":%s", authToken))
	}
	args = append(args, fmt.Sprintf("containers-storage:%s", imageTag), fmt.Sprintf("docker://%s", imageTag))

	// Use skopeo to push the image to the registry
	pushCmd := exec.Command("skopeo", args...)
	pushOutput, pushErr := pushCmd.CombinedOutput()
	if pushErr != nil {
		return fmt.Errorf("push image: %w\n%s", pushErr, pushOutput)
	}
	return nil
}
