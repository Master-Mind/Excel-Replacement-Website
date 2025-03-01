package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Master-Mind/Excel-Replacement-Website/data_loaders"
	"github.com/Master-Mind/Excel-Replacement-Website/templs"

	"github.com/a-h/templ"
)

func loadDataHandler(w http.ResponseWriter, r *http.Request) {
	file, fileheader, err := r.FormFile("file")

	if err != nil {
		errstr := fmt.Sprintf("Error getting file: %v", err)
		fmt.Printf("%s", errstr)
		http.Error(w, errstr, http.StatusInternalServerError)
		return
	}

	defer file.Close()

	if strings.Contains(strings.ToLower(fileheader.Filename), "run") {
		data, err := data_loaders.LoadRunsSpreadsheet(file, 2021)

		if err != nil {
			errstr := fmt.Sprintf("Error loading data: %v", err)
			fmt.Printf("%s", errstr)
			http.Error(w, errstr, http.StatusInternalServerError)
			return
		}

		comp := templs.RunCSVDisplay(data)
		comp.Render(r.Context(), w)
	} else {
		data, err := data_loaders.LoadWeightsSpreadsheet(file, 2022)

		if err != nil {
			errstr := fmt.Sprintf("Error loading data: %v", err)
			fmt.Printf("%s", errstr)
			http.Error(w, errstr, http.StatusInternalServerError)
			return
		}

		comp := templs.LiftCSVDisplay(data)
		comp.Render(r.Context(), w)
	}
}

func main() {
	home_comp := templs.Home()

	http.Handle("/", templ.Handler(home_comp))
	http.HandleFunc("/load-data", loadDataHandler)

	server := &http.Server{Addr: ":80"}

	// Goroutine to listen for shutdown signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\nShutting down server...")
		if err := server.Close(); err != nil {
			fmt.Printf("Error shutting down server: %v\n", err)
		}
	}()

	fmt.Println("Server is listening at port 80...")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Printf("Server error: %v\n", err)
	}
}
