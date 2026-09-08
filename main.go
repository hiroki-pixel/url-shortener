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

// POST usersハンドラー関数の設定
func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost { //ここ必要ないか？Routerへの登録でメソッドも明示しているし
		//後で機能を実装
		log.Println("/usersへのアクセス確認が完了しました")
	} else {
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
	}
}

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
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB接続先の環境変数を確認してください。")
	}

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("DB Open不可")
	}
	if err := db.Ping(); err != nil {
		log.Fatal("DB Connect不可")
	}
	fmt.Println("DB接続完了")

	//Routerの作成
	mux := http.NewServeMux()

	//作成したRouter(mux)にhandlerを登録する
	mux.HandleFunc("POST /users", createUser)

	//設定したアドレスとRouterを使用してサーバを起動
	addr := ":" + port
	log.Fatal(http.ListenAndServe(addr, mux))

}
