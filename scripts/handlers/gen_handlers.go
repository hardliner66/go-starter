// +build ignore

// This program generates handlers.go. It can be invoked by running go generate ./...
// or via go run scripts/handlers/gen_handlers.go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"allaboutapps.at/aw/go-mranftl-sample/internal/util"
	"allaboutapps.at/aw/go-mranftl-sample/scripts/scriptsutil"
)

// https://blog.carlmjohnson.net/post/2016-11-27-how-to-use-go-generate/

var (
	HANDLERS_PACKAGE = "/internal/api/handlers"

	PROJECT_ROOT       = util.GetProjectRootDir()
	PATH_MOD_FILE      = PROJECT_ROOT + "/go.mod"
	PATH_HANDLERS_FILE = PROJECT_ROOT + HANDLERS_PACKAGE + "/handlers.go"

	// <METHOD_PREFIXES>*<METHOD_SUFFIX> must look like this
	// TODO: also check fn signature
	METHOD_PREFIXES = []string{
		"Get", "Head", "Patch", "Post", "Put", "Delete",
	}
	METHOD_SUFFIX = "Route"

	PACKAGE_TEMPLATE = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by scripts/handlers_gen.go
package handlers

import (
	"allaboutapps.at/aw/go-mranftl-sample/internal/api"
	{{- range .SubPkgs }}
	"{{ $.BasePkg }}{{ . }}"
	{{- end }}
)

func AttachAllRoutes(s *api.Server) {
	// attach our routes
	{{- range .Funcs }}
	{{ .PackageName }}.{{ .FunctionName }}(s)
	{{- end }}
}
`))
)

type ResolvedFunction struct {
	PackageName  string
	FunctionName string
}

// get all functions in above handler packages
// that match Get*, Put*, Post*, Patch*, Delete*
func main() {

	baseModuleName, err := scriptsutil.GetModuleName(PATH_MOD_FILE)

	if err != nil {
		log.Fatal(err)
	}

	handlersBasePackage := baseModuleName + HANDLERS_PACKAGE + "/"

	subPkgs := []string{}

	files, err := ioutil.ReadDir(PROJECT_ROOT + HANDLERS_PACKAGE + "/")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		// fmt.Println(f.Name())

		if f.IsDir() {
			subPkgs = append(subPkgs, f.Name())
		}
	}

	funcs := []ResolvedFunction{}

	for _, subPackageName := range subPkgs {

		set := token.NewFileSet()
		packs, err := parser.ParseDir(set, PROJECT_ROOT+HANDLERS_PACKAGE+"/"+subPackageName, nil, 0)

		if err != nil {
			fmt.Println("Failed to parse package:", err)
			os.Exit(1)
		}

		for _, pack := range packs {
			for _, f := range pack.Files {
				for _, d := range f.Decls {

					if fn, isFn := d.(*ast.FuncDecl); isFn {

						fnName := fn.Name.String()

						for _, prefix := range METHOD_PREFIXES {
							if strings.HasPrefix(fnName, prefix) && strings.HasSuffix(fnName, METHOD_SUFFIX) {
								funcs = append(funcs, ResolvedFunction{
									FunctionName: fnName,
									PackageName:  subPackageName,
								})
							}
						}
					}
				}
			}
		}
	}

	// debug print out
	// for _, function := range funcs {
	// 	fmt.Println(function.PackageName, function.FunctionName)
	// }

	f, err := os.Create(PATH_HANDLERS_FILE)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	PACKAGE_TEMPLATE.Execute(f, struct {
		BasePkg string
		SubPkgs []string
		Funcs   []ResolvedFunction
	}{
		BasePkg: handlersBasePackage,
		SubPkgs: subPkgs,
		Funcs:   funcs,
	})

}