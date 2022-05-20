package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Environment variables
const (
	env_port  = "DashboardPort"
	env_trUrl = "TestRunnerUrl"
	env_logs  = "LogFile"
)

// Defaults
const (
	defaultPort = "80"
)

func main() {
	file := redirectLogs()
	defer file.Close()
	serve()
}

// redirectLogs redirects logs to the file specified by the log file environment variable
func redirectLogs() *os.File {
	lf := os.Getenv(env_logs)
	if lf != "" {
		file, err := os.Create(lf)
		if err != nil {
			log.Println(fmt.Errorf("unable to redirect logs to %v\n%w", lf, err))
			return nil
		}
		log.SetOutput(file)
		return file
	}
	return nil
}

// getPort determines which port to use
func getPort() string {
	port := os.Getenv(env_port)
	if port == "" {
		port = defaultPort
	}
	return port
}

// serve creates an HTTP server to host the dashboard
func serve() {
	http.HandleFunc("/", rootHandler)

	addr := ":" + getPort()
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
	// Get test runner url
	trURL := os.Getenv(env_trUrl)
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
