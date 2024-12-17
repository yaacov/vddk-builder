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

func LoadConfig() *Config {
	return &Config{
		ImageName:     getEnv("IMAGE_NAME", "vddk"),
		CAPublicKey:   getEnv("CA_PUBLIC_KEY", "/etc/tls/server.crt"),
		PrivateKey:    getEnv("PRIVATE_KEY", "/etc/tls/server.key"),
		ServerPort:    getEnv("SERVER_PORT", "8443"),
		RegistryUser:  os.Getenv("REGISTRY_USER"),
		RegistryPass:  os.Getenv("REGISTRY_PASS"),
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
