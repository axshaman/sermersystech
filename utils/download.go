package utils

import (
	"net/http"
	"path/filepath"
	"encoding/json"
)

// SendJSONError returns a JSON-formatted error
func SendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// DownloadHandler serves files for download
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" {
		SendJSONError(w, "Missing 'file' parameter", http.StatusBadRequest)
		return
	}
	filePath, err := filepath.Abs(file)
	if err != nil {
		SendJSONError(w, "Invalid file path", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)
}