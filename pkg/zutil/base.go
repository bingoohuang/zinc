package zutil

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// GetEnvInt returns the value of the environment variable named by the key and returns the default value if the environment variable is not set.
func GetEnvInt(key string, fallback int) int {
	if s, _ := os.LookupEnv(key); s != "" {
		if i, err := strconv.ParseInt(s, 10, 32); err == nil {
			return int(i)
		}

		log.Printf("failed to parse env %s as int", s)
	}

	return fallback
}

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
			log.Fatalf("stat dir  %s error: %v", dir, err)
		}

		log.Printf("start to create dir %s", dir)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			log.Printf("W! MkdirAll dir %s, error: %v", dir, err)
			if errors.Is(err, os.ErrPermission) {
				chmodRecursively(dir, 0o777)
			}
		}
	}

	return dir
}

func chmodRecursively(parent string, mode os.FileMode) {
	_, err := os.Stat(parent)
	if err == nil {
		return
	}
	log.Printf("W! stat dir %s, error: %v", parent, err)
	chmodRecursively(filepath.Dir(parent), mode)
	if err := os.Chmod(parent, mode); err != nil {
		log.Printf("W! chmod dir %s, error: %v", parent, err)
	}
}

func SliceContains(slice []string, key string) bool {
	for _, element := range slice {
		if element == key {
			return true
		}
	}

	return false
}
