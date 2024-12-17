package builder

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"vddk-builder/pkg/config"
)

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

func BuildAndPushImage(cfg *config.Config, filePath string, imageName string) {
	// Create ./tmp directory for temporary files
	tmpDir := filepath.Join(".", "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		fmt.Printf("Failed to create temporary directory: %v\n", err)
		return
	}

	// Define the extracted directory under tmp
	extractedDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractedDir, 0755); err != nil {
		fmt.Printf("Failed to create extraction directory: %v\n", err)
		return
	}

	// Defer cleanup for extractedDir and tar.gz file
	defer func() {
		fmt.Println("Cleaning up...")
		if err := os.RemoveAll(extractedDir); err != nil {
			fmt.Printf("Failed to remove extracted directory: %v\n", err)
		}
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove tar.gz file: %v\n", err)
		}
	}()

	// Extract the tar.gz file
	fmt.Println("Extracting uploaded file...")
	if err := extractTarGz(filePath, extractedDir); err != nil {
		fmt.Printf("Failed to extract archive: %v\n", err)
		return
	}

	// Set image name and tag
	if imageName == "" {
		imageName = cfg.ImageName
	}
	imageTag := fmt.Sprintf("%s/%s:latest", cfg.ImageRegistry, imageName)

	// Build the image using podman
	fmt.Println("Building image...")
	cmd := exec.Command("podman", "build", "-f", "Containerfile.vddk", "-t", imageTag, extractedDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to build image: %v\n%s\n", err, output)
		return
	}

	/// Push the image to the registry
	fmt.Println("Pushing image...")
	pushCmd := exec.Command("podman", "push", "--tls-verify=false", imageTag)
	pushOutput, pushErr := pushCmd.CombinedOutput()
	if pushErr != nil {
		fmt.Printf("Failed to push image: %v\n%s\n", pushErr, pushOutput)
		return
	}

	fmt.Println("Image build and push completed successfully.")
}
