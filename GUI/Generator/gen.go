// +build ignore
// This program generates html.go/css.go. It can be invoked by running
// This package is way to have the best performance aspect, however the `go generate` should only run one time,
// so I choose to have a more readable file. The `go generate` creates the code, so is important to anyone understand
// what generated.
// go generate
package main

import (
	"path/filepath"
	"strings"
	"io/ioutil"
	"os"
	"github.com/kib357/less-go"
	"text/template"
	"github.com/brokenbydefault/Nanollet/Config"
	"github.com/brokenbydefault/Nanollet/Util"
)

func main() {
	generateSciter()
	generateLESS()
	generateHTML()
	generateCSS()
}

//@TODO Use template/text instead and rewrite the code
//@TODO Support Linux/Darwin

var sciterTemplate = template.Must(template.New("").Parse(`// +build {{.OS}}
// Code generated by go generate; DO NOT EDIT.
package Front

var Sciter = []byte{ {{- range .Data }}0x{{printf "%X" .}},{{- end }} }
`))

type sciterStruct struct {
	OS   string
	File string
	Data []byte
}

func generateSciter() {
	structs := []sciterStruct{
		{"darwin", "sciter-osx-64.dylib", nil},
		{"windows", "sciter.dll", nil},
		{"linux", "libsciter-gtk-64.so", nil},
	}

	for _, strc := range structs {
		strc.Data, _ = ioutil.ReadFile(strc.File)

		if strc.Data == nil {
			panic("invalid sciter linking")
		}

		store, _ := os.Create("GUI/Front/sciter_" + strc.OS + ".go")
		sciterTemplate.Execute(store, strc)
	}
}

func generateLESS() {
	err := less.RenderFile("GUI/Front/less/style.less", "GUI/Front/css/style.css", map[string]interface{}{"compress": true})
	if err != nil {
		panic(err)
	}
}

var cssTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
package Front

{{if .IsDebug}}
import "github.com/brokenbydefault/Nanollet/Util"
{{end}}

var CSSStyle = {{.Data}}
`))

type cssStruct struct {
	IsDebug bool
	Data    string
}

func generateCSS() {
	strc := cssStruct{}
	strc.IsDebug = Config.IsDebugEnabled()

	// If debug is enable the Nanollet will read the file directly, making possible to change the HTML/CSS without
	// need to `go generate` again.
	if strc.IsDebug {
		strc.Data = `Util.FileToString("GUI/Front/css/style.css")`
	} else {
		strc.Data = "`" + Util.FileToString("GUI/Front/css/style.css") + "`"
	}

	store, _ := os.Create("GUI/Front/css.go")
	cssTemplate.Execute(store, strc)
}

var htmlTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
package Front

{{if .IsDebug}}
import "github.com/brokenbydefault/Nanollet/Util"
{{end}}

type HTMLPAGE string

{{- range .HTML}}
var HTML{{.Name}} = HTMLPAGE({{.Data}})
{{- end}}
`))

type pageStruct struct {
	Data string
	Name string
}
type htmlStruct struct {
	IsDebug bool
	HTML    []pageStruct
}

func generateHTML() {
	strc := htmlStruct{}
	strc.IsDebug = Config.IsDebugEnabled()

	pages, _ := filepath.Glob("GUI/Front/html/*")
	for _, path := range pages {
		page := pageStruct{}
		page.Name = strings.Title(strings.Replace(filepath.Base(path), ".html", "", 1))

		// If debug is enable the Nanollet will read the file directly, making possible to change the HTML/CSS without
		// need to `go generate` again.
		if strc.IsDebug {
			page.Data = `Util.FileToString("` + filepath.ToSlash(path) + `")`
		} else {
			page.Data = "`" + Util.FileToString(path) + "`"
		}

		strc.HTML = append(strc.HTML, page)
	}

	store, _ := os.Create("GUI/Front/html.go")
	htmlTemplate.Execute(store, strc)
}
