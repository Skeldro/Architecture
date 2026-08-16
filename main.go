package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var listTmpl = template.Must(template.New("list").Parse(`<!doctype html>
<title>Collab Docs</title>
<h1>Collab Docs</h1>
<form method="POST" action="/create">
  <input name="title" placeholder="New document title" autofocus>
  <button>Create</button>
</form>
{{if .}}<ul>
{{range .}}  <li><a href="/doc?name={{.}}">{{.}}</a></li>
{{end}}</ul>{{else}}<p>No documents yet.</p>{{end}}`))

var editorTmpl = template.Must(template.New("editor").Parse(`<!doctype html>
<title>{{.Name}} — Collab Docs</title>
<p><a href="/">&larr; all documents</a></p>
<h1>{{.Name}}</h1>
<form method="POST" action="/save">
  <input type="hidden" name="name" value="{{.Name}}">
  <textarea name="content" rows="25" cols="90">{{.Content}}</textarea>
  <br>
  <button>Save</button>
</form>`))

func main() {
	os.MkdirAll("docs", 0755)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		entries, err := os.ReadDir("docs")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var titles []string
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".txt") {
				titles = append(titles, strings.TrimSuffix(e.Name(), ".txt"))
			}
		}
		listTmpl.Execute(w, titles)
	})

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		title := r.FormValue("title")
		if title == "" || filepath.Base(title) != title || title[0] == '.' {
			http.Error(w, "bad title: must be a plain file name", 400)
			return
		}
		path := "docs/" + title + ".txt"
		if _, err := os.Stat(path); err == nil {
			http.Error(w, "a document with that title already exists", 409)
			return
		}
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, "/doc?name="+title, http.StatusSeeOther)
	})

	http.HandleFunc("/doc", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		if name == "" || filepath.Base(name) != name || name[0] == '.' {
			http.Error(w, "bad name", 400)
			return
		}
		content, err := os.ReadFile("docs/" + name + ".txt")
		if err != nil {
			http.Error(w, "no such document", 404)
			return
		}
		editorTmpl.Execute(w, map[string]string{"Name": name, "Content": string(content)})
	})

	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		if name == "" || filepath.Base(name) != name || name[0] == '.' {
			http.Error(w, "bad name", 400)
			return
		}
		// browsers submit textarea content with \r\n; D5 says files hold plain \n
		content := strings.ReplaceAll(r.FormValue("content"), "\r\n", "\n")
		if err := os.WriteFile("docs/"+name+".txt", []byte(content), 0644); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, "/doc?name="+name, http.StatusSeeOther)
	})

	fmt.Println("Collab Docs running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
