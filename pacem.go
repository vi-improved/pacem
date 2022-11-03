package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Score struct {
	KeriLee string
	Caroline string
	Horn string
	Sax string
	Bass string
}

var inputs = make(map[string]int)
var generated bool = false

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

var sections = [2]string{"intro", "test"}

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
	for _, sectiondir := range sections {
		concatenatedPart.WriteString(partReader(sectiondir, part))
	}
	return concatenatedPart.String()
}

func scoreGenerator() {
	p := Score {
		KeriLee: partConcatenator("KeriLee"),
		Caroline: partConcatenator("Caroline"),
		Horn: partConcatenator("Horn"),
		Sax: partConcatenator("Sax"),
		Bass: partConcatenator("Bass"),
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
	if part == "" {
		http.Redirect(w, r, "/input/", http.StatusFound)
		return
	}
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
	"<input type=\"radio\" id=\"kerilee\" name=\"performer\" value=\"Keri Lee\">"+
	"<label for=\"kerilee\">Keri Lee</label><br>"+
	"<input type=\"radio\" id=\"caroline\" name=\"performer\" value=\"Carolyn\">"+
	"<label for=\"caroline\">Caroline</label><br>"+
	"<input type=\"radio\" id=\"horn\" name=\"performer\" value=\"Horn\">"+
	"<label for=\"horn\">Horn</label><br>"+
	"<input type=\"radio\" id=\"sax\" name=\"performer\" value=\"Sax\">"+
	"<label for=\"sax\">Sax</label><br>"+
	"<input type=\"radio\" id=\"bass\" name=\"performer\" value=\"Bass\">"+
	"<label for=\"bass\">Bass</label><br>"+
	"<input type=\"submit\" value=\"Save\">"+
	"</form>")
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	choice, err := strconv.Atoi(r.FormValue("body"))
	performer := r.FormValue("performer")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	inputs[performer] = choice
	generated = false
	fmt.Print(inputs)
	http.Redirect(w, r, "/wait/" + strings.ToLower(strings.ReplaceAll(performer, " ", "")), http.StatusFound)
}

func waitHandler(w http.ResponseWriter, r *http.Request) {
	performer := r.URL.Path[len("/wait/"):]
	if !generated {
		fmt.Fprintf(w, "<head><meta http-equiv=\"refresh\" content=\"5\" /></head><body>Number submitted, waiting for other performers</body>")
	} else {
		http.Redirect(w, r, "/" + performer, http.StatusFound)
	}
}

func waitToGenerate() {
	for true {
		for len(inputs) < 5 {
			time.Sleep(1 * time.Second)
		}
		var seed int64 = 0
		for _, v := range inputs {
			v = int(math.Pow(float64(v), math.Mod(float64(v), 11)))
			seed += int64(v)
		}
		rand.Seed(seed)
		scoreGenerator()
		inputs = make(map[string]int)
		generated = true
	}
}

func main() {
	http.HandleFunc("/", partHandler)
	http.HandleFunc("/submit/", submitHandler)
	http.HandleFunc("/input/", inputHandler)
	http.HandleFunc("/wait/", waitHandler)
	go waitToGenerate()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
