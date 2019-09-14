package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

/*
Page struct describes how page data is stored in memory.
The Body is in type 'byte' because of the ioutil work we
will be doing
*/
type Page struct {
	Title string
	Body  []byte // Byte slice.
}

/*
Cache template to reduce inefficiencies when calling renderTemplate

template.Must is a wrapper that panics when passed a non-nil error
value, otherwise it returns the *Template unaltered. A panic is
appropriate here; if the templates can't be loaded, the only
sensible thing to do is exit the program
*/
var templates = template.Must(
	template.ParseFiles("edit.html", "view.html"))

/*
Disallow invalid path names (e.g. ../) to be viewed/edited on the
server's file system.

regexp.MustCompile will parse and compile the regexp and return
regextp.Regexp. Note, MustCompile is distinct from Compile in that
it will panic if the expression compilation fails, while Compile
*/
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

/*
getTitle uses the validPath regexp to validate the path and extract
the page title. If title is valid, it will return with a nil error. If
the title is invalid, the function will return a 400 status code to the
HTTP connection and return a error to the handler

errors.New allows you to write your own errors with custom messages
*/
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil
}

/*
save function takes in a pointer to a Page, writes the file
and returns a value of type error (because that's the return
type of ioutil.Writefile). This method saves the Page's Body
to a text file. To keep it simple, make Title the file's name.

The 0600 used in the WriteFile function is octal flag for
read and write (rw) permissions.
*/
func (p *Page) save() error {
	filename := p.Title + ".txt"
	fmt.Println("saving" + filename)
	return ioutil.WriteFile(filename, p.Body, 0600)
}

/*
loadPage constructs file name from title parameter, reads
the file's contents into a new variable, body and returns
a ptr to Page literal constructed with the proper title + body
values and also an error (if thrown).

Callers of this function can check 2nd parameter, if it is nil,
then it has successfully loaded a Page, otherwise an error was
thrown – which can be handled accordingly
*/
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

/*
templates.ExecuteTemplate executes the parsed cached template,
and writes the generated HTML to the http.ResponseWriter. If there
is an error whilst parsing the file, then a http.Error is thrown
with a 500 status code to indicate internal server error
*/
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/*
viewHandler allows you to view a wiki page. If it doesn't exist
It redirects the client to edit the Page so the content may be
created.

http.Redirect adds an HTTP status code of http.StatusFound (302)
and a Location header to the HTTP response.
*/
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("view: " + r.URL.Path)
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	}
	renderTemplate(w, "view", p)
}

/*
editHandler loads the page, (or if it doesn't exist, creates an
empty Page struct) and then displayes a HTML form.
*/
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		// if page doesn't exist, create a new one with the
		// Page struct
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

/*
saveHandler handles submission of forms located on the edit pages.

StatusInternalServerError is thrown if file cannot be saved. This
is so that any errors that occur during p.save() are reported to
the user.
*/
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	fmt.Println("save: " + r.URL.Path)
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

/*
makeHandler takes a function of the type (http.ResponseWriter,
r *http.Request) and returns a function of type http.HandlerFunc. This
allows you to repeat the validation/error checking for each of the end
points at once. (Function literals and closures)

The closure returned by makeHandler is a function that takes an
http.ResponseWriter and *http.Request (i.e. an http.HandlerFunc).
The closure extracts the title from the request path and valides it with
the regexp. If the title is invalid, an error will be written to the
ResponseWriter using http.NotFound. If the title is valid, the enclosed
handler function fn will be called with the ResponseWriter, Request and
title as arguments
*/
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
