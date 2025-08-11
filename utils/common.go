package utils

import (
	"math/rand"
	"time"
	"database/sql"
	"fmt"
	"strings"
)

func GenerateRandomPassword(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func NullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func GetObjectTypeID(db *sql.DB, name string) (int64, error) {
	var id int64
	err := db.QueryRow("SELECT id FROM objects_types WHERE name = ?", strings.ToLower(name)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("object_type '%s' not found", name)
	}
	return id, nil
}
