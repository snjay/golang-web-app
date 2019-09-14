package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
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
t.ParseFiles reads the contents of edit.html file and
returns a ptr to a template.Template. If there is an error whilst
parsing the file, then a http.Error is thrown with a 500 status
code to indicate internal server error

t.Execute executes the template, writing the generated HTML to the
http.ResponseWriter.
*/
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles(tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, p)
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
func viewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("view: " + r.URL.Path)
	title := r.URL.Path[len("/view/"):]
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	}
	renderTemplate(w, "view", p)
}

/*
editHandler loads the page, (or if it doesn't exist, creates an empty
Page struct) and then displayes a HTML form.
*/
func editHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("edit: " + r.URL.Path)
	title := r.URL.Path[len("/view/"):]
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
func saveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("save: " + r.URL.Path)
	title := r.URL.Path[len("/save/"):]
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
