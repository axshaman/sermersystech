package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// SecurityLogEntry represents the structure of a security event
type SecurityLogEntry struct {
	Event     string `json:"event"`
	JWTUserID int    `json:"jwt_user_id"`
	ReqUserID int    `json:"req_user_id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Path      string `json:"path"`
	Timestamp string `json:"timestamp"`
}

// LogCSHandler receives JSON log events and prints/stores them
func LogCSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var entry SecurityLogEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)

	// Log to stdout
	log.Printf("🚨 [SECURITY LOG] event=%s jwt_user_id=%d req_user_id=%d ip=%s path=%s ua=%s",
		entry.Event, entry.JWTUserID, entry.ReqUserID, entry.IP, entry.Path, entry.UserAgent)

	// Save to ./security_logs/YYYY-MM-DD_HHMMSS_<event>.json
	logDir := "security_logs"
	_ = os.MkdirAll(logDir, 0755)

	filename := fmt.Sprintf("%s_%s.json", time.Now().Format("2006-01-02_150405"), entry.Event)
	filePath := filepath.Join(logDir, filename)

	f, err := os.Create(filePath)
	if err == nil {
		defer f.Close()
		json.NewEncoder(f).Encode(entry)
	} else {
		log.Printf("❌ Failed to write incident file: %v", err)
	}

	// Optional: also log to a central file
	f2, err := os.OpenFile("security_events.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f2.Close()
		json.NewEncoder(f2).Encode(entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "logged"})
}