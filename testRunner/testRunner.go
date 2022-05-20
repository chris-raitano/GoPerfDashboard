package main

import (
	"goperfordashboard/testrunner/env"
	"goperfordashboard/testrunner/requesthandlers"
	"log"
	"net/http"
	"os"
)

// Defaults
const (
	defaultPort = "80"
)

func main() {
	file := configureLogs()
	defer file.Close()

	startService()
}

// configureLogs onfigures the log package output
func configureLogs() *os.File {
	// Display line numbers
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// Redirect to log file provided in env variable
	lf := os.Getenv(env.LOG_FILE)
	if lf != "" {
		file, err := os.Create(lf)
		if err != nil {
			log.Printf("Unable to redirect logs to %v\n%v\n", lf, err)
			return nil
		}
		log.SetOutput(file)
		return file
	}
	return nil
}

// getPort determines which port to use
func getPort() string {
	port := os.Getenv(env.PORT)
	if port == "" {
		port = defaultPort
	}
	return port
}

// startService creates an HTTP server to recieve test requests
func startService() {
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
		requesthandlers.PostModuleUpload(w, r)
	default:
		w.WriteHeader(http.StatusForbidden)
	}
}
