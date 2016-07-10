package main

import (
    "log"
    "net/http"
    "io"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, "hello, world\n")
}

func main() {

    // public views
    http.HandleFunc("/", HandleIndex)

    log.Fatal(http.ListenAndServe(":8080", nil))
}
