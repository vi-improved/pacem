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
	"strings"
	"text/template"
	"time"
)

type Score struct {
	KeriLee  string
	Caroline string
	Horn     string
	Sax      string
	Bass     string
}

var templates = template.Must(template.ParseFiles("template.ly"))

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

func scoreGenerator() error {
	p := Score{
		KeriLee:  partConcatenator("KeriLee"),
		Caroline: partConcatenator("Caroline"),
		Horn:     partConcatenator("Horn"),
		Sax:      partConcatenator("Sax"),
		Bass:     partConcatenator("Bass"),
	}
	var filledTemplate bytes.Buffer
	templates.ExecuteTemplate(&filledTemplate, "template.ly", p)
	out := filledTemplate.Bytes()
	os.WriteFile("out.ly", out, 0600)
	lilypond := exec.Command("lilypond", "out.ly")
	return lilypond.Run()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/pdf")
	err := scoreGenerator()
	if err != nil {
		return
	}
	out, _ := os.ReadFile("out.pdf")
	io.WriteString(w, string(out))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
