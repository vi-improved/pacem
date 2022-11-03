package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
)

type Score struct {
	KeriLee  string
	Caroline string
	Horn     string
	Sax      string
	Bass     string
}

var templates = template.Must(
	template.ParseFiles(
		"scoreTemplate.ly",
		"kerileeTemplate.ly",
		"carolineTemplate.ly",
		"hornTemplate.ly",
		"saxTemplate.ly",
		"bassTemplate.ly",
	),
)

func partReader(title string, part string) string {
	entries := 0
	dirlist, _ := os.ReadDir("sections/" + title)
	for _, direntry := range dirlist {
		if strings.HasPrefix(direntry.Name(), part) {
			entries++
		}
	}
	randentry := rand.Intn(entries)
	result, _ := os.ReadFile("sections/" + title + "/" + part + fmt.Sprint(randentry) + ".ly")
	return string(result)
}

func partConcatenator(part string) string {
	var concatenatedPart strings.Builder
	sectiondirs, _ := os.ReadDir("sections")
	for _, sectiondir := range sectiondirs {
		concatenatedPart.WriteString(partReader(sectiondir.Name(), part))
	}
	return concatenatedPart.String()
}

func scoreGenerator() {
	p := Score{
		KeriLee:  partConcatenator("KeriLee"),
		Caroline: partConcatenator("Caroline"),
		Horn:     partConcatenator("Horn"),
		Sax:      partConcatenator("Sax"),
		Bass:     partConcatenator("Bass"),
	}
	parts := [6]string{"score", "kerilee", "caroline", "horn", "sax", "bass"}
	for _, part := range parts {
		var filledPartTemplate bytes.Buffer
		templates.ExecuteTemplate(&filledPartTemplate, part+"Template.ly", p)
		partOut := filledPartTemplate.Bytes()
		os.WriteFile(part+".ly", partOut, 0600)
		partLilypond := exec.Command("lilypond", part+".ly")
		partLilypond.Run()
	}
}

func partHandler(w http.ResponseWriter, r *http.Request) {
	part := r.URL.Path[1:]
	w.Header().Add("Content-Type", "application/pdf")
	out, err := os.ReadFile(part + ".pdf")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(out))
}

func inputHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Submit Number</h1>"+
	"<form action=\"/submit/\" method=\"POST\">"+
	"<textarea name=\"body\"></textarea><br>"+
	"<input type=\"submit\" value=\"Save\">"+
	"</form>")
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	choice, err := strconv.Atoi(r.FormValue("body"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rand.Seed(int64(choice))
	scoreGenerator()
	io.WriteString(w, "number submitted and score randomized")
}

func main() {
	http.HandleFunc("/", partHandler)
	http.HandleFunc("/submit/", submitHandler)
	http.HandleFunc("/input/", inputHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
