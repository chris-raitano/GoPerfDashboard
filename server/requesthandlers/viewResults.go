package requesthandlers

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"
)

type resultsViewTmplParams struct {
	Coverage []byte
	Mem      []byte
	Cpu      []byte
}

// viewResultsPage loads the test results page
func GetViewResultsPage(w http.ResponseWriter, r *http.Request, res []byte) {
	var params resultsViewTmplParams
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
