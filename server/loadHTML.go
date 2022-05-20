package main

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"
)

const (
	indexTemplateFile   = "templates/index.html.tmpl"
	resultsTemplateFile = "templates/view-results.html.tmpl"
	resultsCSSFile      = "templates/view-results.css.tmpl"
	resultsJSFile       = "templates/view-results.js.tmpl"
)

type uploadPageParams struct {
	ErrEl string
}

// fileUploadPage loads the go module upload page
func fileUploadPage(w http.ResponseWriter, r *http.Request, errMsg string) {
	err := template.Must(template.ParseFiles(indexTemplateFile)).Execute(w, uploadPageParams{ErrEl: errMsg})
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type resultsViewParams struct {
	Coverage []byte
	Mem      []byte
	Cpu      []byte
}

// viewResultsPage loads the test results page
func viewResultsPage(w http.ResponseWriter, r *http.Request, res []byte) {
	var params resultsViewParams
	err := json.Unmarshal(res, &params)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := template.Must(template.ParseFiles(resultsTemplateFile, resultsCSSFile, resultsJSFile)).ExecuteTemplate(w, "results", params); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
