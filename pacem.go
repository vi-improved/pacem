package main

import (
	"fmt"
	"math/rand"
	"text/template"
	"time"
)

var templates = template.Must(template.ParseFiles("intro.ly"))

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Print(rand.Intn(10))
}
