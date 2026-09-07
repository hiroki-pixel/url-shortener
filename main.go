package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
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

	//サーバの起動確認
	addr := ":" + port
	log.Fatal(http.ListenAndServe(addr, nil))

}
