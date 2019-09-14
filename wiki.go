package main

import (
	"fmt"
	"io/ioutil"
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
values and also an error (if thrown). Callers of this function
can check 2nd parameter, if it is nil, then it has successfully
loaded a Page, otherwise an error was thrown – which can be
handled
*/
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample page")}
	p1.save()
	p2, _ := loadPage("TestPage")
	fmt.Println(string(p2.Body))
}
