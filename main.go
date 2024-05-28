package main

import (
	"fmt"
	"log"
	"net/http"

	never "never/HTML"
)

func main() {
	http.HandleFunc("/", never.HandleRequest)
	http.HandleFunc("/artist", never.HandleRequest2)
	fmt.Println("go to -->  http://localhost:5500/ ")
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates/"))))
	http.Handle("/HTML/", http.StripPrefix("/HTML/", http.FileServer(http.Dir("HTML/"))))
	log.Fatal(http.ListenAndServe(":5500", nil))
	

	if err := http.ListenAndServe(":5500", nil); err != nil {
		log.Fatal(err)
	}
}
