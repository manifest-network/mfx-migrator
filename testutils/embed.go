package testutils

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func CopyDirFromEmbedFS(sourceFS fs.FS, sourceDir, targetDir string) error {
	return fs.WalkDir(sourceFS, sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, path)

		if d.IsDir() {
			return os.MkdirAll(targetPath, os.ModePerm)
		}

		// For files, create file in the target path
		file, err := sourceFS.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Create target file
		targetFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		// Copy the contents
		_, err = io.Copy(targetFile, file)
		return err
	})
}
