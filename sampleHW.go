package main

import (
	"fmt"
	"net/http"
)

func drafthello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprint(res, "Hello, World and Max!")
}

func draftmain() {
	http.HandleFunc("/", hello)
	http.ListenAndServe("localhost:4000", nil)

}
