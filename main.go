package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	//環境変数の取得確認
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error Loading .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not found")
	}
	log.Println("PORT:", port)

	//DBへの接続確認
	db_url := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", db_url)
	if db_url == "" {
		log.Fatal("DB接続先の環境変数を確認してください。")
	}
	if err != nil {
		log.Fatal("DB Open不可")
	}
	if err := db.Ping(); err != nil {
		log.Fatal("DB Connect不可")
	}
	fmt.Println("DB接続完了")

	//サーバの起動確認
	addr := ":" + port
	log.Fatal(http.ListenAndServe(addr, nil))

}
