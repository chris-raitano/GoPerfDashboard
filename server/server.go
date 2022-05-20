package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	port    = 8082
	testURL = "http://testrunner-service:8081"
)

func main() {
	// Send logs to file
	file, _ := os.Create("out.log")
	defer file.Close()
	log.SetOutput(file)

	serve()
}

// serve creates an HTTP server to host the dashboard
func serve() {
	http.HandleFunc("/", rootHandler)

	addr := fmt.Sprintf(":%v", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Unable to start web server. %v\n", err.Error())
	}
}

// rootHandler handles requests to the root url.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		uploadHandler(w, r)
	case "GET":
		fileUploadPage(w, r, "")
	}
}

// uploadHandler handles file upload requests.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Forward request to test runner
	req, err := http.NewRequest("POST", testURL, r.Body)
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
	viewResultsPage(w, r, b)
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
	fileUploadPage(w, r, htmlMsg)
}
