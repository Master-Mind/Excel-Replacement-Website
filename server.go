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

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path)

	if r.URL.Path == "/" {
		dbhandling.WorkoutHandler(w, r)
		return
	}
}

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

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/runs", dbhandling.RunHandler)
	http.HandleFunc("/run-page", dbhandling.RunPage)
	http.HandleFunc("/trans-data", dbhandling.TransformData)
	http.Handle("/import", templ.Handler(templs.Import()))
	http.HandleFunc("/new-run", dbhandling.AddRun)
	http.HandleFunc("/delete-run", dbhandling.RemoveRun)
	http.HandleFunc("/new-shoe", dbhandling.NewShoe)
	http.HandleFunc("/delete-shoe", dbhandling.DeleteShoe)
	http.HandleFunc("/delete-set", dbhandling.DeletSet)
	http.HandleFunc("/delete-workout", dbhandling.DeleteWorkout)
	http.HandleFunc("/new-workout", dbhandling.CreateWorkout)
	http.HandleFunc("/new-set", dbhandling.AddSet)
	http.HandleFunc("/diet", dbhandling.DietPageHandler)
	http.HandleFunc("/transform-nut", dbhandling.AddNutritionData)
	http.HandleFunc("/recommend-food", dbhandling.FoodRecomendationHandler)

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
