package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Master-Mind/Excel-Replacement-Website/dbhandling"
	"github.com/Master-Mind/Excel-Replacement-Website/templs"
	"github.com/a-h/templ"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	err = dbhandling.InitDB()

	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}

	http.HandleFunc("/", dbhandling.WorkoutHandler)
	http.HandleFunc("/runs", dbhandling.RunHandler)
	http.HandleFunc("/trans-data", dbhandling.TransformData)
	http.Handle("/import", templ.Handler(templs.Import()))
	http.HandleFunc("/new-run", dbhandling.AddRun)
	http.HandleFunc("/delete-run", dbhandling.RemoveRun)
	http.HandleFunc("/new-shoe", dbhandling.NewShoe)
	http.HandleFunc("/delete-shoe", dbhandling.DeleteShoe)

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
