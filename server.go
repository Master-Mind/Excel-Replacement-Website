package main

import (
	"fmt"
	"net/http"

	"github.com/Master-Mind/Excel-Replacement-Website/templs"

	"github.com/a-h/templ"
)

func main() {
	home_comp := templs.Home()

	http.Handle("/", templ.Handler(home_comp))

	fmt.Println("Server is listening at port 80...")
	http.ListenAndServe(":80", nil)
}
