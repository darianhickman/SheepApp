package main

import (
	"io"
	"html/template"
	"encoding/csv"
	"github.com/golang/glog"
	"os"
	"strings"
	"io/ioutil"
	"net/http"
	"bytes"
	"bufio"
)

func main() {
	//Keep the static functions in a separate file
	//if len(os.Args) < 1{
	//	os.Args = append(os.Args, "index.html")
	//}
	//if len(os.Args) < 2{
	//	os.Args = append(os.Args, "appsheet.csv")
	//}

	tpl := SetupTemplate("index.html")
	data := SetupCSV("appsheet.csv")
	f, err := os.Create("/tmp/dat2.html")
	check(err)
	w := bufio.NewWriter(f)
	//defer f.Close()

	GenApp(w, tpl, data, nil)
}


func check(e error) {
	if e != nil {
		panic(e)
	}
}

func SetupCSV(source string) csv.Reader{
	//accomodate file, url, string
	// check file exists
	var reader = csv.NewReader(nil)
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		f, err := os.Open(source)
		if err != nil { glog.Fatalf("Failed Open: ", source, err)}
		//defer f.Close() // this needs to be after the err check
		reader = csv.NewReader(f)
		glog.Info("file:", source, reader)
		return *reader
	}
	if strings.Index("http:", source[0:10]) > -1{
		resp, err := http.Get(source)
		if err != nil {
			glog.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil{ glog.Fatal(err)}
		glog.Info("string", source, reader)
		reader = csv.NewReader(bytes.NewReader(body))

	}else{
		reader = csv.NewReader(strings.NewReader(source))
	}

	glog.Fatal("Failed to load: ", source)
	return *reader
}

func SetupTemplate(source string) (template.Template){
	//accomodate file, url, string
	// first assuming the string passed in is the template content itself
	var htmltpl = template.Must((template.New("itml").Parse(source)))

	// check file exists then use
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		// since the source filepath exists then use it as ref to template file.
		htmltpl = template.Must(template.New("itml").ParseFiles(source))
		glog.Info("file: ", source)

		return *htmltpl
	}
	if strings.Index("http", source[0:10]) > -1{
		// since the source starts with http then use it as url linking to template file.
		resp, err := http.Get(source)
		if err != nil {
			// handle error
			glog.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil{ glog.Fatal(err)}

		htmltpl = template.Must(template.New("itml").Parse(string(body[:])))
		glog.Info("http", source, htmltpl.DefinedTemplates())

	}else{

		glog.Info("string", source, htmltpl.DefinedTemplates())

	}

	glog.Fatal("Failed to load: ", source)

	return *htmltpl
}

type App struct {
	Style, Bodies, Texts []map[string]string
}

func GenApp (w io.Writer, tpl template.Template, reader csv.Reader, schema []string ){
	// not sure how to make the schema dynamic hmm.
	// well it would be great if template needed to change but this injection functions didn't need an update.

	header, err := reader.Read()
	glog.Info("header:", header)
	glog.Info("template" , tpl.DefinedTemplates())
	if err != nil {
		glog.Fatal(err)
	}
	hm := make(map[string]int)

	//assume header column present
	for idx, val := range header {
		hm[val] = idx
	}

	style := make([]map[string]string, 1, 1)
	bodies := make([]map[string]string, 1, 1)
	texts := make([]map[string]string, 1, 1)



	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if record[hm["type"]] == "css" {
			css := make(map[string]string)

			css["selector"] = record[hm["selector"]]
			css["key"] = record[hm["key"]]
			css["value"] = record[hm["value"]]
			style = append(style, css)

		}
		if record[hm["type"]] == "html" {
			html := make(map[string]string)

			html["selector"] = record[hm["selector"]]
			html["key"] = record[hm["key"]]
			html["value"] = record[hm["value"]]
			bodies = append(bodies, html)
		}
		if record[hm["type"]] == "text" {
			text := make(map[string]string)

			text["selector"] = record[hm["selector"]]
			text["key"] = record[hm["key"]]
			text["value"] = record[hm["value"]]
			texts = append(texts, text)
		}

	}
	docgut := App{Style: style,
		Bodies: bodies,
		Texts: texts,}

	tpl.Execute(w, docgut)
	glog.Info(w)
}