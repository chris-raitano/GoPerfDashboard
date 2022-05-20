package test

import (
	"goperfordashboard/testrunner/fileops"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	testFileName = "testFile.generated.go"
	templateFile = "templates/testgen.go.tmpl"
	testModDir   = "testMod"
	profDataDir  = "perfdashboard/data"
	outDir       = "perfdashboard/formattedresults"
)

type outFiles struct {
	CoverProf string
	CoverHtml string
	MemProf   string
	MemSvg    string
	CpuProf   string
	CpuSvg    string
}

type results struct {
	Coverage []byte
	Mem      []byte
	Cpu      []byte
}

type templParams struct {
	ModName      string
	CPUTestCount int
	Files        outFiles
}

// GenPerfReports runs tests and generates the output files
func GenPerfReports() (results, error) {
	fpaths := outFiles{
		CoverProf: filepath.Join(profDataDir, "cover.out"),
		CoverHtml: filepath.Join(outDir, "cover.html"),
		MemProf:   filepath.Join(profDataDir, "mem.out"),
		MemSvg:    filepath.Join(outDir, "mem.svg"),
		CpuProf:   filepath.Join(profDataDir, "cpu.out"),
		CpuSvg:    filepath.Join(outDir, "cpu.svg"),
	}

	tmpl := template.Must(template.ParseFiles(templateFile))

	// cd into module directory
	restoreDir, err := os.Getwd()
	if err != nil {
		log.Printf("Unable to find current directory. %v\n", err)
		return results{}, err
	}
	defer os.Chdir(restoreDir)
	os.Chdir(testModDir)

	// Create directories for output
	if err := createOutputDirs(); err != nil {
		log.Printf("Unable to create output directories. %v\n", err)
	}

	// Run tests
	if err := genTestData(fpaths, tmpl); err != nil {
		log.Printf("An error occurred running tests\n%v\n", err)
		return results{}, err
	}
	// Write results
	rslt, err := readResultsFiles(fpaths)
	if err != nil {
		log.Printf("Unable to read test results\n%v\n", err)
	}

	return rslt, err
}

// createOutputDirs creates directory structure to store output files
func createOutputDirs() error {
	if err := os.MkdirAll(profDataDir, os.ModePerm /*777*/); err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, os.ModePerm /*777*/); err != nil {
		return err
	}

	return nil
}

// genTestData enters the test diretory, generates a test file and executes the tests
func genTestData(fpaths outFiles, tmpl *template.Template) error {

	// Get module name
	testMod, err := getModName()
	if err != nil {
		log.Printf("Unable to find module. %v\n", err)
		return err
	}

	// Create test file to run
	if err = writeTestFile(tmpl, testMod, fpaths); err != nil {
		log.Printf("Unable to create test file. %v\n", err)
		return err
	}

	// Run tests
	if err = runTests(); err != nil {
		log.Printf("Unable to run tests. %v\n", err)
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

// readResultsFiles copies perf report files into a result object
func readResultsFiles(fpaths outFiles) (results, error) {
	// Define files to read
	f := map[string]string{
		"coverage": fpaths.CoverHtml,
		"memory":   fpaths.MemSvg,
		"cpu":      fpaths.CpuSvg,
	}
	rslt := make(map[string][]byte)

	// Read all files
	for k, v := range f {
		bytes, err := fileops.Readf(v)
		if err != nil {
			log.Printf("Unable to read %v file. %v", k, err)
			return results{}, err
		}
		rslt[k] = bytes
	}

	return results{
		Coverage: rslt["coverage"],
		Mem:      rslt["memory"],
		Cpu:      rslt["cpu"],
	}, nil
}

// Runs tests for the test module
func runTests() error {
	return exec.Command("go", "generate").Run()
}
