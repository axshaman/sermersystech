package utils

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB() *sql.DB {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("❌ MYSQL_DSN не установлен в переменных окружения")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к базе данных: %v", err)
	}

	log.Println("✅ Подключено к базе данных через MYSQL_DSN")
	return db
}
