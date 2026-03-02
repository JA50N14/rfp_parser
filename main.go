package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/JA50N14/rfp_parser/config"
	"github.com/JA50N14/rfp_parser/walk"
	"github.com/joho/godotenv"
)

type TimerPayload struct {
	Data struct {
		ScheduleStatus struct {
			Last string `json:"Last"`
			Next string `json:"Next"`
		} `json:"ScheduleStatus"`
		IsPastDue bool `json:"IsPastDue"`
	} `json:"Data"`
}

func main() {
	port := os.Getenv("FUNCTIONS_CUSTOMHANDLER_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", timerHandler)

	log.Printf("Custom Handler listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func timerHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload TimerPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Println("Failed to parse JSON:", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Timer fired! Last: %s, Next: %s, PastDue: %t",
		payload.Data.ScheduleStatus.Last,
		payload.Data.ScheduleStatus.Next,
		payload.Data.IsPastDue,
	)

	err := runParser(r.Context())
	if err != nil {
		log.Println("Error ocurred parsing packages", err)
		http.Error(w, "FAILED", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func runParser(ctx context.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if os.Getenv("ENV") == "local" {
		if err := godotenv.Load(".env"); err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	cfg, err := config.NewApiConfig(logger)
	if err != nil {
		return fmt.Errorf("failed to initialize API config: %w", err)
	}

	err = walk.WalkDocLibrary(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to walk document library: %w", err)
	}

	cfg.Logger.Info("RFP Package(s) successfully processed!")
	return nil
}
