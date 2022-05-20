package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	maxUploadSize = 50 << 20 // 50 MB
	testModDir    = "testMod"
	testFileName  = "testFile.generated.go"
	templateFile  = "templates/testgen.go.tmpl"
	outputDir     = "perfdashboard/out"
	htmlDir       = "perfdashboard/html"
	port          = 8081
)

type outFiles struct {
	CoverProf string
	CoverHtml string
	MemProf   string
	MemSvg    string
	CpuProf   string
	CpuSvg    string
}

type templParams struct {
	ModName      string
	CPUTestCount int
	Files        outFiles
}

func main() {
	// Send logs to file
	file, _ := os.Create("out.log")
	defer file.Close()
	log.SetOutput(file)

	startService()
}

type resultsHTML struct {
	Coverage []byte
	Mem      []byte
	Cpu      []byte
}

// startService creates an HTTP server to recieve test requests
func startService() {
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
		fpaths := outFiles{
			CoverProf: filepath.Join(outputDir, "cover.out"),
			CoverHtml: filepath.Join(htmlDir, "cover.html"),
			MemProf:   filepath.Join(outputDir, "mem.out"),
			MemSvg:    filepath.Join(htmlDir, "mem.svg"),
			CpuProf:   filepath.Join(outputDir, "cpu.out"),
			CpuSvg:    filepath.Join(htmlDir, "cpu.svg"),
		}

		// Retrieve the file from form data
		r.ParseMultipartForm(maxUploadSize)
		file, header, err := r.FormFile("upload")
		if err != nil {
			handleUploadError("Unable to get file from request body.", err, w, r)
			return
		}
		defer file.Close()
		// Extract file to disk
		if err := extractFile(file, header); err != nil {
			handleUploadError("Unalbe to extract file to disk.", err, w, r)
			return
		}
		// Run tests
		if err := genTestData(fpaths); err != nil {
			handleUploadError("An error occurred running tests.", err, w, r)
			return
		}
		// Write results
		rslt, err := readResultsHtml(fpaths)
		if err != nil {
			handleUploadError("Unable to read test results.", err, w, r)
		}
		msg, err := json.Marshal(rslt)
		if err != nil {
			handleUploadError("Unable to marshall results into response.", err, w, r)
		}
		w.Header().Set("Content-Type", "application/text")
		w.Write(msg)
	default:
		w.WriteHeader(http.StatusForbidden)
	}
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
	if err := copyFile(file, header); err != nil {
		log.Printf("Unable to create local copy of uploaded file.")
		return err
	}

	// Extract zip file
	if err := extractZip(header.Filename); err != nil {
		log.Printf("An error occurred extracting the zip file.")
		return err
	}
	return nil
}

func handleUploadError(msg string, err error, w http.ResponseWriter, r *http.Request) {
	errStr := fmt.Errorf("%v %w", msg, err)
	log.Println(errStr)
}

// genTestData enters the test diretory, generates a test file and executes the tests
func genTestData(fpaths outFiles) error {
	tmpl := template.Must(template.ParseFiles(templateFile))

	restoreDir, err := os.Getwd()
	if err != nil {
		log.Printf("Unable to find current directory. %v\n", err.Error())
		return err
	}
	defer os.Chdir(restoreDir)
	os.Chdir(testModDir)

	testMod, err := getModName()
	if err != nil {
		log.Printf("Unable to find module. %v\n", err.Error())
		return err
	}

	if err = writeTestFile(tmpl, testMod, fpaths); err != nil {
		log.Printf("Unable to create test file. %v\n", err.Error())
		return err
	}

	if err = createOutputDirs(); err != nil {
		log.Printf("Unable to create output directories. %v\n\n", err.Error())
		return err
	}

	if err = runTests(); err != nil {
		log.Printf("Unable to run tests. %v\n", err.Error())
		return err
	}

	return nil
}

// getModName searches the current directory for a go module and returns the name
func getModName() (string, error) {
	out, err := exec.Command("go", "list", "-m").Output()
	return strings.Trim(string(out), "\n\t "), err
}

// writeTestFile creates a test file and populates it using a template
func writeTestFile(tmpl *template.Template, modName string, fpaths outFiles) error {
	f, err := os.Create(testFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, templParams{
		ModName:      modName,
		CPUTestCount: 1000,
		Files:        fpaths,
	})
}

// createOutputDirs creates directory structure to store output files
func createOutputDirs() error {
	if err := os.MkdirAll(outputDir, os.ModePerm /*777*/); err != nil {
		return err
	}
	if err := os.MkdirAll(htmlDir, os.ModePerm /*777*/); err != nil {
		return err
	}

	return nil
}

// readResultsHtml reads results files and builds a result object
func readResultsHtml(fpaths outFiles) (resultsHTML, error) {
	cov, err := readf(filepath.Join(".", "testMod", fpaths.CoverHtml))
	if err != nil {
		log.Printf("Unable to read coverage file. %v", err)
		return resultsHTML{}, err
	}
	mem, err := readf(filepath.Join(".", "testMod", fpaths.MemSvg))
	if err != nil {
		log.Printf("Unable to read memory file. %v", err)
		return resultsHTML{}, err
	}
	cpu, err := readf(filepath.Join(".", "testMod", fpaths.CpuSvg))
	if err != nil {
		log.Printf("Unable to read cpu file. %v", err)
		return resultsHTML{}, err
	}

	return resultsHTML{
		Coverage: cov,
		Mem:      mem,
		Cpu:      cpu,
	}, nil
}

// Runs tests for the test module
func runTests() error {
	return exec.Command("go", "generate").Run()
}
