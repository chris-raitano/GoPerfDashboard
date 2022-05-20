package requesthandlers

import (
	"encoding/json"
	"errors"
	"goperfordashboard/testrunner/fileops"
	"goperfordashboard/testrunner/test"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

const (
	maxUploadSize = 50 << 20 // 50 MB
)

// PostModuleUpload handles module uploads and generates reports for the tests
func PostModuleUpload(w http.ResponseWriter, r *http.Request) {
	// Retrieve the file from form data
	r.ParseMultipartForm(maxUploadSize)
	file, header, err := r.FormFile("upload")
	if err != nil {
		log.Printf("Unable to get file from request body\n%v\n", err)
		return
	}
	defer file.Close()
	// Extract file to disk
	if err := extractFile(file, header); err != nil {
		log.Printf("Unalbe to extract file to disk\n%v\n", err)
		return
	}

	// Run tests and generate reports
	rslt, err := test.GenPerfReports()
	if err != nil {
		log.Printf("Unable to generate reports.\n%v\n", err)
	}

	// Return data as json
	msg, err := json.Marshal(rslt)
	if err != nil {
		log.Printf("Unable to marshall results into response\n%v\n", err)
	}
	w.Header().Set("Content-Type", "application/text")
	w.Write(msg)
}

// extractFile reads a .zip file from the request and extracts it onto the local filesystem
func extractFile(file multipart.File, header *multipart.FileHeader) error {
	// Only accept zip files
	if ext := filepath.Ext(header.Filename); ext != ".zip" {
		err := errors.New("invalid filetype")
		log.Printf("Invalid filetype uploaded: %v.", ext)
		return err
	}

	// Copy to filesystem
	if err := fileops.CopyFile(file, header); err != nil {
		log.Printf("Unable to create local copy of uploaded file.")
		return err
	}

	// Extract zip file
	if err := fileops.ExtractZip(header.Filename); err != nil {
		log.Printf("An error occurred extracting the zip file.")
		return err
	}
	return nil
}
