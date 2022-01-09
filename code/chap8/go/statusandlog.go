package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func status(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Unsupported Method.", http.StatusNotFound)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(">***** Status Start *****")
	fmt.Printf("%s\n", body)
	fmt.Println("<***** Status End   *****")
}

func logs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Unsupported Method.", http.StatusNotFound)
		return
	}

	reader, err := gzip.NewReader(r.Body)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	fmt.Println(">***** DecisionLog Start *****")
	fmt.Printf("%s\n", body)
	fmt.Println("<***** DecisionLog End   *****")
}

func main() {
	http.HandleFunc("/status", status)
	http.HandleFunc("/logs", logs)

	fmt.Printf("Starting server at port 9111\n")
	log.Fatal(http.ListenAndServe(":9111", nil))
}
