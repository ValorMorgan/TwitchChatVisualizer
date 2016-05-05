package main

/*===================================================
	Imports
=====================================================
- Name			Purpose
=====================================================
- fmt			Handles print statements
- log
- strings		Functionalities for Strings
- encoding/json
- io
- io/ioutil		Handles io between fileSize
- html/template	Handles storing HTML in outside file
- net/http		Handles web-application listening
				capabilities
===================================================*/
import (
	"fmt"
	"log"
	"strings"
	"encoding/json"
	"io"
	"io/ioutil"
	"html/template"
	"net/http"
)

/*===================================================
	Globals
===================================================*/
var TEMPLATES = template.Must(template.ParseFiles("index.html","NotFound.html"))

/*===================================================
	Custom Objects
===================================================*/
/*==================
	Page
==================*/
type Page struct {
	Title	string
	Body	[]byte
}
func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

/*==================
	appError
==================*/
type appError struct {
    Error   error
    Message string
    Code    int
}

/*==================
	appHandler
==================*/
type appHandler func(http.ResponseWriter, *http.Request) *appError
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if e := fn(w, r); e != nil {
        http.Error(w, e.Message, e.Code)
    }
}

/*==================
	Message
==================*/
type Message struct {
	Name, Text string
}

/*===================================================
	Functions
===================================================*/
func loadPage(file string) (*Page, error) {
	body, err := ioutil.ReadFile(file)
    return &Page{Title: strings.Split(file, ".")[0], Body: body}, err
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := TEMPLATES.ExecuteTemplate(w, tmpl, p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func requestData(w http.ResponseWriter, strm string) ([]string, error) {
	req, err := http.NewRequest("GET", strm, nil)
	if err != nil {
		panic(err)
    }

	var messages []string
	dec := json.NewDecoder(req.Body)
	for {
		var m Message
		err := dec.Decode(&m)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		
		messages = append(messages, fmt.Sprintf("%s: %s\n", m.Name, m.Text))
	}
	
	return messages, nil
}

func openStreamConnection(w http.ResponseWriter, strm string) (string, error) {
	//server := "irc.chat.twitch.tv"
	//port := 6667
	sInfo, err := requestData(w, "https://api.twitch.tv/kraken/channels/"+strm)
	if err != nil {
		return "", err
    }
	//cInfo, err := requestData(w, "https://api.twitch.tv/kraken/channels/"+strm)
	//if err != nil {
    //    return err
    //}
	
	p1 := &Page{Title: "StreamInfo", Body: []byte(strings.Join(sInfo,"\n"))}
	p1.save()
	p2, _ := loadPage("StreamInfo.txt")
	fmt.Println(string(p2.Body))
	
	return strings.Join(sInfo,"\n"), nil
}

/*===================================================
	Handlers
===================================================*/
func handler(w http.ResponseWriter, r *http.Request) *appError {
	title := strings.TrimSpace(r.URL.Path[1:])
	if (title == "" || len(title) < 1) {
		title = "index.html"
	} else if !(strings.Contains(title, ".")) {
		title = title + ".html"
	}
	p, err := loadPage(title)
    if err != nil {
        return &appError{err, "Page not found", 404}
    }
	renderTemplate(w, title, p)
	
	return nil
}

func streamerHandler(w http.ResponseWriter, r *http.Request) *appError {
    title := strings.TrimSpace(r.URL.Path[len("/streamer/"):])
	if strings.TrimSpace(title) == "" {
		return &appError{nil, "Page not found", 404}
	}
	body, err := openStreamConnection(w, title)
	if err != nil {
        return &appError{err, "Can't access stream", 500}
    }
    fmt.Fprintf(w, "<p>%s</p>", body)
	
	return nil
}

/*===================================================
	Main
===================================================*/
func main() {
    http.Handle("/", appHandler(handler))
	http.Handle("/streamer/", appHandler(streamerHandler))
    log.Fatal(http.ListenAndServe(":8080", nil))
}