package zutil

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// GetEnv returns the value of the environment variable named by the key and returns the default value if the environment variable is not set.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetS3Bucket() string {
	bucket := GetEnv("S3_BUCKET", "")
	return bucket
}

func GetDataDir() string {
	pre, err := os.UserHomeDir()
	if err != nil {
		log.Printf(" os.UserHomeDir() error: %v", err)
		pre = "."
	}

	dir := GetEnv("ZINC_DIR", filepath.Join(pre, ".zinc/data"))
	if _, err := os.Stat(dir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal().Msgf("failed to stat dir %s: %v", dir, err)
		}
		if err := os.MkdirAll(dir, 0777); err != nil {
			if errors.Is(err, os.ErrPermission) {
				chmodRecursively(dir, 0777)
			}
		}
	}

	return dir
}

func chmodRecursively(parent string, mode os.FileMode) {
	if _, err := os.Stat(parent); err == nil {
		return
	}

	chmodRecursively(filepath.Dir(parent), mode)
	os.Chmod(parent, mode)
}

func SliceContains(slice []string, key string) bool {
	for _, element := range slice {
		if element == key {
			return true
		}
	}

	return false
}
