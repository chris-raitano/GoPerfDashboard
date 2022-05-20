package requesthandlers

import (
	"fmt"
	"goperfordashboard/webserver/env"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
)

type uploadPageTmplParams struct {
	ErrEl string
}

// GetUploadPage loads the go module upload page
func GetUploadPage(w http.ResponseWriter, r *http.Request, errMsg string) {
	err := template.Must(template.ParseFiles(indexTemplateFile)).Execute(w, uploadPageTmplParams{ErrEl: errMsg})
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// PostUpload handles file upload requests.
func PostUpload(w http.ResponseWriter, r *http.Request) {
	// Get test runner url
	trURL := os.Getenv(env.TEST_RUNNER_URL)
	if trURL == "" {
		panic("Test Runner URL not set. Can't run tests.")
	}

	// Forward request to test runner
	req, err := http.NewRequest("POST", trURL, r.Body)
	if err != nil {
		handleUploadError("Unable to create test request.", err, w, r)
	}
	req.Header = r.Header
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		handleUploadError("Unable to run tests.", err, w, r)
	}

	// Redirect to results page on success
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Unable to read test result. %v", err)
	}
	GetViewResultsPage(w, r, b)
}

// handleUploadError handles file upload errors and reloads the indexc page
func handleUploadError(msg string, err error, w http.ResponseWriter, r *http.Request) {
	htmlMsg := fmt.Sprintf("%v %v\n", msg, err)
	if htmlMsg == "" {
		htmlMsg = "An unknown error occurred"
	}
	// Format message
	htmlMsg = fmt.Sprintf("<div class=\"err-msg\">%v</div>", htmlMsg)

	errStr := fmt.Errorf("%v %w", msg, err)
	log.Println(errStr)
	// reload upload page on error
	GetUploadPage(w, r, htmlMsg)
}
