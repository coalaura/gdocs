package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)

	for range ticker.C {
		err := cleanup()
		if err != nil {
			log.Warnf("cleanup: %v\n", err)
		}
	}
}

func cleanup() error {
	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		return nil
	}

	now := time.Now()

	return filepath.WalkDir("docs", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if now.Sub(info.ModTime()) > time.Hour {
			log.Printf("cleaned %q\n", filepath.Base(path))

			return os.Remove(path)
		}

		return nil
	})
}
