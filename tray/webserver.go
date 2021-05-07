package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	
)


func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, Configuration.Name)
}

func landingPage(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(landingTemplate()))
}

func public_info(w http.ResponseWriter, req *http.Request) {
	out, _ := json.Marshal(Info)
	fmt.Fprintf(w, string(out))
}

func webserver() {
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/", landingPage)
	http.HandleFunc("/public_info", public_info)
	http.ListenAndServe(fmt.Sprintf(":%v", Configuration.HttpPort), nil)
}



func landingTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>UMH</title>
  </head>
  <body>
<h1>UMH Menu</h1>
Web service interface to UMH menu.

<a href="/upload">Upload</a> a file, or view the <a href="/public_info">details</a> for this computer
  </body>
</html>`

}
