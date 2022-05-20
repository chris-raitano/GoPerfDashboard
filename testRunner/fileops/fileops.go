package fileops

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// copyFile copies a file to the current directory.
func CopyFile(file multipart.File, header *multipart.FileHeader) error {
	dstFile, err := os.Create(header.Filename)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy filecontents to filesystem
	if _, err := io.Copy(dstFile, file); err != nil {
		return err
	}
	return nil
}

// extractZip extracts a .zip file in the current directory
func ExtractZip(fname string) error {
	zipFile, err := zip.OpenReader(fname)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	for _, f := range zipFile.File {
		if err := extractFileFromZip(f); err != nil {
			return err
		}
	}

	return nil
}

// extractFileFromZip extracts a single file from a zipped directory
func extractFileFromZip(f *zip.File) error {
	fpath, err := filepath.Abs(filepath.Join(".", f.Name))
	if err != nil {
		return err
	}

	// Check for zip slip attack
	fpath = filepath.Clean(fpath)
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(fpath, currentDir) {
		return fmt.Errorf("invalid file name: %v", f.Name)
	}

	// Create directory
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(fpath, f.Mode()); err != nil {
			return err
		}
		return nil
	}

	// Create parent dirs
	if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm /*777*/); err != nil {
		return err
	}

	// Create file
	dstFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Unzip and copy contents
	zipFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zipFile.Close()
	if _, err := io.Copy(dstFile, zipFile); err != nil {
		return err
	}

	return nil
}

// readf reads a single file into a []byte
func Readf(filename string) ([]byte, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return content, nil
}
