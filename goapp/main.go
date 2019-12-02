package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

// MyConfig contains configuration structure
type MyConfig struct {
	Username string
	Start    string
	End      string
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	log.Print("MYAPP, version ", VERSION, " (", BUILDDATE, ")")

	server := http.NewServeMux()

	myconfig := MyConfig{
		Username: getEnv("MYAPP_USERNAME", "myuser"),
		Start:    getEnv("MYAPP_START", "mystart"),
		End:      getEnv("MYAPP_END", "myend"),
	}

	// simple api: getting {"content": "some text"}, it returns {"result": "texts some"}
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("myapp invoked - ", r.Method)

		tmpl, err := template.ParseFiles("static/hello.html")

		if err != nil {
			log.Println("unable to parse html file")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "unable to parse html file")
		} else {
			err = tmpl.Execute(w, myconfig)
			if err != nil {
				log.Println("unable to execute template")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "unable to execute template")
			}
		}
	})

	// listining
	port := getEnv("PORT", "8001")
	log.Print("Running server at :", port)

	log.Fatal(http.ListenAndServe(":"+port, server))
}
