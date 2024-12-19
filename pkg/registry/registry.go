package registry

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

// CheckImageExists checks if the specified image exists in the registry.
func CheckImageExists(imageName, registryURL, authToken string) (bool, error) {
	// Construct the image manifest URL
	url := fmt.Sprintf("https://%s/v2/%s/manifests/latest", registryURL, imageName)

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
