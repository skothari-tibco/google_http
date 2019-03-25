//+build ignore
package main

import (
	"errors"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/project-flogo/cli/util"
)

var (
	shimDir        string
	newShimSupport string
	oldShimSupport string
	newEmbedded    string
	oldEmbedded    string
	newImports     string
	oldImports     string
	modFile        string
	buildShimFile  string
)

func main() {
	fmt.Println("Executing build script for Google-http Trigger ...")

	pwd, err := os.Getwd()

	fmt.Println("Current Dir is ...", pwd)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error Encountered.. %v\n", err)
		os.Exit(1)
	}

	setVars(pwd)

	err = createInitShim()

	if err != nil {
		restoreImportsFile()
		fmt.Fprintf(os.Stderr, "Error Encountered in initialization of shim.. %v\n", err)
		os.Exit(1)
	}

	err = startFunc(pwd)

	if err != nil {
		restoreImportsFile()
		fmt.Fprintf(os.Stderr, "Error Encountered in starting function.. %v\n", err)
		os.Exit(1)
	}
}

func setVars(pwd string) {

	shimDir = "shim"
	modFile = filepath.Join(pwd, "go.mod")

	newShimSupport = filepath.Join(pwd, shimDir, "shim_support.go")
	oldShimSupport = filepath.Join(pwd, "shim_support.go")

	newEmbedded = filepath.Join(pwd, shimDir, "embeddedapp.go")
	oldEmbedded = filepath.Join(pwd, "embeddedapp.go")

	newImports = filepath.Join(pwd, shimDir, "imports.go")
	oldImports = filepath.Join(pwd, "imports.go")

	buildShimFile = filepath.Join(pwd, "build.go")

}

func createInitShim() error {
	var err error

	err = copyAndSet(oldShimSupport, newShimSupport)
	if err != nil {
		return err
	}
	err = copyAndSet(oldImports, newImports)
	if err != nil {
		return err
	}
	err = copyAndSet(oldEmbedded, newEmbedded)
	if err != nil {
		return err
	}
	//Rename build.go file to avoid confusion in gcloud tool...
	err = os.Rename(buildShimFile, buildShimFile+".bak")

	if err != nil {
		return err
	}
	return nil
}

func copyAndSet(oldFile, newFile string) error {

	var err error

	err = os.MkdirAll(filepath.Dir(newFile), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.Rename(oldFile, newFile)
	if err != nil {
		return err
	}

	err = setPkg(newFile)

	if err != nil {
		return err
	}

	return nil
}

func setPkg(file string) error {

	read, err := ioutil.ReadFile(file)

	var data string
	if strings.Contains(string(read), "package main") {
		data = strings.Replace(string(read), "package main", "package shim", -1)
	} else {
		data = strings.Replace(string(read), "module main", "module shim", -1)
	}
	err = ioutil.WriteFile(file, []byte(data), 0644)

	if err != nil {
		return err
	}
	return nil
}

func startFunc(pwd string) error {

	shimFile := filepath.Join(pwd, "shim.go")

	//Change mod main to mod shim in go.mod
	setPkg(modFile)

	//Add the " _ shim/shim" imports to shim.go
	addShimImportToFile(shimFile, "shim/shim")
	cmd := exec.Command("gcloud", "functions", "deploy", "Handle", "--runtime", "go111", "--trigger-http")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf(string(out.Bytes()))
	}

	return nil
}

func addShimImportToFile(shimFile, pkg string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, shimFile, nil, 0)
	if err != nil {
		return err
	}

	if !util.AddImport(fset, file, pkg) {
		return errors.New("Error in adding package to shim file")
	}

	f, err := os.Create(shimFile)
	defer f.Close()
	if err := printer.Fprint(f, fset, file); err != nil {
		return err
	}
	return nil
}

func restoreImportsFile() {
	_ = os.Rename(newImports, oldImports)

}
