package registry

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
)

// CheckImageExists checks if a Docker image exists in the specified registry.
// It sends a HEAD request to the image manifest URL and checks the HTTP status code.
//
// Parameters:
//   - imageName: The name of the Docker image to check.
//   - registryURL: The URL of the Docker registry.
//   - authToken: The authentication token for the registry (optional).
//
// Returns:
//   - bool: True if the image exists, false otherwise.
//   - error: An error if the request fails or an unexpected status code is returned.
func CheckImageExists(imageName, registryURL, authToken string) (bool, error) {
	// Split image name into name and tag
	name, tag := splitImageName(imageName)

	// Construct the image manifest URL
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registryURL, name, tag)

	// Create a new HTTP request
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err
	}

	// Set Authorization header if needed
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	// Set Accept header to request image manifest, including OCI support
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")

	// Send the HTTP request
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode == http.StatusOK {
		return true, nil // Image exists
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil // Image does not exist
	}

	return false, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
}

// splitImageName splits the image name into name and tag.
// If no tag is provided, it defaults to "latest".
func splitImageName(imageName string) (string, string) {
	parts := strings.Split(imageName, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], "latest"
}
