package main

import (
	"goperfdashboard/webserver/env"
	"goperfdashboard/webserver/requesthandlers"
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
	serve()
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

// serve creates an HTTP server to host the dashboard
func serve() {
	http.HandleFunc("/", rootHandler)

	addr := ":" + getPort()
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("unable to start web server\n%v\n", err)
	}
}

// rootHandler handles requests to the root url.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		requesthandlers.PostUpload(w, r)
	case "GET":
		requesthandlers.GetUploadPage(w, r, "")
	}
}
