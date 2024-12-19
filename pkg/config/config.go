package config

import "os"

type Config struct {
	ImageName     string
	CAPublicKey   string
	PrivateKey    string
	ServerPort    string
	RegistryUser  string
	RegistryPass  string
	UploadDir     string
	ImageRegistry string
}

// LoadConfig loads the configuration for the application from environment variables.
// It returns a pointer to a Config struct populated with the following fields:
// - ImageName: The name of the image, defaults to "vddk" if not set.
// - CAPublicKey: The path to the CA public key, defaults to "/etc/tls/server.crt" if not set.
// - PrivateKey: The path to the private key, defaults to "/etc/tls/server.key" if not set.
// - ServerPort: The port on which the server will run, defaults to "8443" if not set.
// - UploadDir: The directory where uploads will be stored, defaults to "/tmp/uploads" if not set.
// - ImageRegistry: The image registry URL, defaults to "image-registry.openshift-image-registry.svc:5000" if not set.
func LoadConfig() *Config {
	return &Config{
		ImageName:     getEnv("IMAGE_NAME", "vddk"),
		CAPublicKey:   getEnv("CA_PUBLIC_KEY", "/etc/tls/server.crt"),
		PrivateKey:    getEnv("PRIVATE_KEY", "/etc/tls/server.key"),
		ServerPort:    getEnv("SERVER_PORT", "8443"),
		UploadDir:     getEnv("UPLOAD_DIR", "/tmp/uploads"),
		ImageRegistry: getEnv("IMAGE_REGISTRY", "image-registry.openshift-image-registry.svc:5000"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
