package requesthandlers

import (
	"encoding/json"
	"errors"
	"goperfdashboard/testrunner/fileops"
	"goperfdashboard/testrunner/test"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

// PostModuleUpload handles module uploads and generates reports for the tests
func PostModuleUpload(w http.ResponseWriter, r *http.Request) {

	// Extract file to disk
	fname := r.Header.Get("filename")
	if fname == "" {
		log.Println("Unable to read file name")
		return
	}
	if err := extractFile(r.Body, fname); err != nil {
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
func extractFile(file io.Reader, filename string) error {
	// Only accept zip files
	if ext := filepath.Ext(filename); ext != ".zip" {
		err := errors.New("invalid filetype")
		log.Printf("Invalid filetype uploaded: %v.", ext)
		return err
	}

	// Copy to filesystem
	if err := fileops.CopyFile(file, filename); err != nil {
		log.Printf("Unable to create local copy of uploaded file.")
		return err
	}

	// Extract zip file
	if err := fileops.ExtractZip(filename); err != nil {
		log.Printf("An error occurred extracting the zip file.")
		return err
	}
	return nil
}
