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
	KeriLee  string
	Carolyn string
	Rachel   string
	Garrett  string
	Anthony  string
}

var (
	inputs         = make(map[string]int)
	generated bool = false
)

var templates = template.Must(
	template.ParseFiles(
		"kerileeTemplate.ly",
		"carolynTemplate.ly",
		"rachelTemplate.ly",
		"garrettTemplate.ly",
		"anthonyTemplate.ly",
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
	p := Score{
		KeriLee:  partConcatenator("KeriLee"),
		Carolyn:  partConcatenator("Carolyn"),
		Rachel:     partConcatenator("Rachel"),
		Garrett:      partConcatenator("Garrett"),
		Anthony:     partConcatenator("Anthony"),
	}
	parts := [5]string{"kerilee", "carolyn", "rachel", "garrett", "anthony"}
	for _, part := range parts {
		var filledPartTemplate bytes.Buffer
		templates.ExecuteTemplate(&filledPartTemplate, part+"Template.ly", p)
		partOut := filledPartTemplate.Bytes()
		os.WriteFile(part+".ly", partOut, 0600)
		partLilypond := exec.Command("lilypond", part+"Part.ly")
		partLilypond.Run()
	}
	scoreLilypond := exec.Command("lilypond", "score.ly")
	scoreLilypond.Run()
}

func partHandler(w http.ResponseWriter, r *http.Request) {
	part := r.URL.Path[1:]
	if part == "" {
		http.Redirect(w, r, "/input/", http.StatusFound)
		return
	}
	w.Header().Add("Content-Type", "application/pdf")
	if part == "score" {
		out, _ := os.ReadFile("score.pdf")
		io.WriteString(w, string(out))
		return
	}
	out, err := os.ReadFile(part + "Part.pdf")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(out))
}

func inputHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := os.ReadFile("input.html")
	fmt.Fprintf(w, string(page))
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
	http.Redirect(w, r, "/wait/"+strings.ToLower(strings.ReplaceAll(performer, " ", "")), http.StatusFound)
}

func waitHandler(w http.ResponseWriter, r *http.Request) {
	performer := r.URL.Path[len("/wait/"):]
	if !generated {
		if len(inputs) < 5 {
			fmt.Fprintf(
				w,
				"<head><meta http-equiv=\"refresh\" content=\"1\" /></head><body>Number submitted, waiting for other performers (%s/5)</body>",
				fmt.Sprint(len(inputs)),
			)
		} else {
			fmt.Fprint(w, "<head><meta http-equiv=\"refresh\" content=\"1\" /></head><body>Number submitted, waiting for score generation</body>")
		}
	} else {
		http.Redirect(w, r, "/"+performer, http.StatusFound)
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
