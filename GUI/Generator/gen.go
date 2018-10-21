// This program generates html.go/css.go. It can be invoked by running
// This package is way to have the best performance aspect, however the `go generate` should only run one time,
// so I choose to have a more readable file. The `go generate` creates the code, so is important to anyone understand
// what generated.

// go generate
package main

import (
	"github.com/brokenbydefault/Nanollet/Util"
	"github.com/kib357/less-go"
	"golang.org/x/net/html"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"github.com/brokenbydefault/Nanollet/Storage"
	"bytes"
)

func main() {
	generateSciter()
	generateLESS()
	generateHTML()
	generateCSS()
}

var sciterTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// +build {{.OS}}

package Front

var Sciter = []byte{ {{.Data}} }
`))

type sciterStruct struct {
	OS   string
	File string
	Data string
}

func generateSciter() {
	structs := []sciterStruct{
		{"darwin", "sciter-osx-64.dylib", ""},
		{"windows", "sciter.dll", ""},
		{"linux", "libsciter-gtk.so", ""},
	}

	for _, strc := range structs {
		bin, _ := ioutil.ReadFile(strc.File)

		hex := Util.UnsafeHexEncode(bin)

		bindata := make([]byte, len(bin)*5)
		isciter := 0
		for i := 0; i < len(hex); i += 2 {
			copy(bindata[isciter:], []byte{0x30, 0x78, hex[i], hex[i+1], 0x2C})
			isciter += 5
		}

		strc.Data = string(bindata)
		if strc.Data == "" {
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
// +build !js

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
	strc.IsDebug = Storage.Configuration.DebugStatus

	// If debug is enable the Nanollet will read the file directly, making possible to change the HTML/CSS without
	// need to `go generate` again.
	if strc.IsDebug {
		strc.Data = `Util.FileToString("GUI/Front/css/style.css")`
	} else {
		strc.Data = "`" + Util.FileToString("GUI/Front/css/style.css") + "`"
	}

	hf, err := os.Create("Nanollet.css")
	if err == nil {
		hf.Write([]byte(Util.FileToString("GUI/Front/css/style.css")))
		hf.Close()
	}

	store, _ := os.Create("GUI/Front/css.go")
	cssTemplate.Execute(store, strc)
}

var htmlTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// +build !js

package Front

{{if .IsDebug}}
import "github.com/brokenbydefault/Nanollet/Util"
{{end}}

var HTML = {{.HTML}}
`))

type htmlStruct struct {
	IsDebug bool
	HTML    string
}

func generateHTML() {
	strc := htmlStruct{}
	strc.IsDebug = Storage.Configuration.DebugStatus

	base, err := os.Open("GUI/Front/html/0_base.html")
	if err != nil {
		panic(err)
	}

	defer base.Close()

	htm, err := html.Parse(base)
	if err != nil {
		panic(err)
	}

	section, ok := getElement(htm, "", []html.Attribute{{
		Key: "class",
		Val: "dynamic",
	}})
	if !ok {
		panic("not found")
	}

	controlBar, ok := getElement(htm, "", []html.Attribute{{
		Key: "class",
		Val: "control",
	}})
	if !ok {
		panic("not found")
	}

	pages, _ := filepath.Glob("GUI/Front/html/*")
	for _, path := range pages {
		if strings.Replace(filepath.Base(path), ".html", "", 1) == "0_base" {
			continue
		}

		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		apphtml, err := html.Parse(file)
		if err != nil {
			panic(err)
		}

		app, ok := getElement(apphtml, "", []html.Attribute{
			{Key: "application"},
		})

		addChild(section[0], app[0])

		// Ignore if the app is not intended to be in the sidebar
		if len(app[0].Attr) > 1 && app[0].Attr[1].Key == "no-sidebar" {
			continue
		}

		// Sidebar button of the APP (e.g "Nanollet", "Nanofy")
		addChild(controlBar[0], &html.Node{
			FirstChild: &html.Node{
				FirstChild: &html.Node{
					Type: html.TextNode,
					Data: strings.Title(app[0].Attr[0].Val),
				},
				Type: html.ElementNode,
				Data: "span",
				Attr: []html.Attribute{{
					Key: "class",
					Val: "title",
				}},
				NextSibling: &html.Node{
					Type: html.ElementNode,
					Data: "span",
					Attr: []html.Attribute{{
						Key: "class",
						Val: "pointer",
					}},
				},
			},
			Type: html.ElementNode,
			Data: "button",
			Attr: []html.Attribute{{
				Key: "id",
				Val: app[0].Attr[0].Val,
			}},
		})

		// Sidebar button for each page of the APP (e.g "Send", "Receive")
		subMenu := &html.Node{
			Type: html.ElementNode,
			Data: "aside",
			Attr: []html.Attribute{{
				Key: "class",
				Val: "application",
			}, {
				Key: "id",
				Val: app[0].Attr[0].Val,
			}},
		}

		s, ok := getElement(app[0], "section", nil)
		if !ok {
			panic("section not found")
		}
		for _, v := range s {
			addChild(subMenu, &html.Node{
				Type: html.ElementNode,
				Data: "button",
				Attr: []html.Attribute{{
					Key: "class",
					Val: strings.Title(v.Attr[0].Val),
				}},
				FirstChild: &html.Node{
					Type: html.ElementNode,
					Data: "span",
					Attr: []html.Attribute{{
						Key: "class",
						Val: "block",
					}},
					FirstChild: &html.Node{
						Type: html.ElementNode,
						Data: "icon",
						Attr: []html.Attribute{{
							Key: "class",
							Val: "icon-" + v.Attr[0].Val,
						}},
						NextSibling: &html.Node{
							FirstChild: &html.Node{
								Type: html.TextNode,
								Data: strings.Title(v.Attr[0].Val),
							},
							Type: html.ElementNode,
							Data: "span",
							Attr: []html.Attribute{{
								Key: "class",
								Val: "title",
							}},
							NextSibling: &html.Node{
								Type: html.ElementNode,
								Data: "span",
								Attr: []html.Attribute{{
									Key: "class",
									Val: "pointer",
								}},
							},
						},
					},
				},
			})
		}

		addChild(controlBar[0], subMenu)

		file.Close()
	}

	b := bytes.NewBuffer(nil)
	html.Render(b, htm)
	strc.HTML = "`" + b.String() + "`"

	hf, err := os.Create("Nanollet.html")
	if err == nil {
		hf.Write(b.Bytes())
		hf.Close()
	}

	store, _ := os.Create("GUI/Front/html.go")
	htmlTemplate.Execute(store, strc)
}

func getElement(node *html.Node, name string, attr []html.Attribute) (resp []*html.Node, ok bool) {
	if node.Data == name || name == "" {
		matches := 0

		if attr != nil {
			for _, v := range node.Attr {
				for _, ex := range attr {
					if v.Key == ex.Key && (v.Val == ex.Val || ex.Val == "") {
						matches += 1
					}
				}
			}
		}

		if matches == len(attr) {
			resp, ok = append(resp, node), true
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if r, find := getElement(c, name, attr); find {
			resp, ok = append(resp, r...), true
		}
	}

	return
}

func addChild(node *html.Node, add *html.Node) {
	if node.FirstChild == nil {
		node.FirstChild = add
		return
	}

	c := node.FirstChild
	for {
		if c.NextSibling == nil {
			c.NextSibling = add
			return
		}
		c = c.NextSibling
	}
}
