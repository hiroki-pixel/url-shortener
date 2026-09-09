package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// 英数字を含むかどうかのバリデーションチェックのための関数
func containsAlphaAndNumber(s string) bool {
	hasAlpha := false
	hasDigit := false

	for _, r := range s {
		switch {
		case 'A' <= r && r <= 'Z':
			hasAlpha = true
		case 'a' <= r && r <= 'z':
			hasAlpha = true
		case '0' <= r && r <= '9':
			hasDigit = true
		}
		if hasAlpha && hasDigit {
			return true
		}
	}
	return false
}

// POST usersハンドラー関数の設定
func createUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	//jsonで受けとったリクエスト情報をGoで扱えるようにDecodeする
	type user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	params := user{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Printf("Failed to Decode Request Body %v", err)
		http.Error(w, "不正なリクエストです\n", http.StatusBadRequest)
		return
	}
	//バリデーションチェックの実装
	//usernameのバリデーションチェック
	if params.Username == "" {
		http.Error(w, "usernameは必須項目です\n", http.StatusBadRequest)
		return
	}

	if len(params.Username) < 5 || len(params.Username) > 15 {
		http.Error(w, "usernameは5桁以上15桁以下で設定する必要があります\n", http.StatusBadRequest)
		return
	}

	//passwordのバリデーションチェック
	if params.Password == "" {
		http.Error(w, "passwordは必須項目です\n", http.StatusBadRequest)
		return
	}

	if len(params.Password) < 8 || len(params.Password) > 64 {
		http.Error(w, "passwordは8桁以上64桁以下で設定する必要があります\n", http.StatusBadRequest)
		return
	}

	if containsAlphaAndNumber(params.Password) == false {
		http.Error(w, "passwordは英数字を両方含む必要があります\n", http.StatusBadRequest)
		return
	}

	log.Printf("usernameは%sで、passwordは%sです\n", params.Username, "隠ぺい")

	//後で書く：処理として、user_idを払い出す。また、パスワードをハッシュ化する。
	userID := uuid.NewString()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Failed to Create HashPassword")
		http.Error(w, "パスワードの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	string(passwordHash)

	//後で書く：リクエスト内容と処理結果をDBに格納する

	//レスポンス情報を作る
	type resuser struct {
		UserID    string `json:"user_id"`
		Username  string `json:"username"`
		CreatedAt string `json:"created_at"`
	}
	resparams := resuser{}

	//created_atの払出(つまりDB格納)がまだなので仮値をハードコーディング
	resparams.UserID = userID
	resparams.CreatedAt = "2026-09-09T12:00:00+09:00"
	resparams.Username = params.Username

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resparams)
	if err != nil {
		log.Printf("Failed to Encode Response %v\n", err)
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
	//mux.HandleFunc("POST /users", createUser)

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		createUser(w, r, db)
	})

	//設定したアドレスとRouterを使用してサーバを起動
	addr := ":" + port
	log.Fatal(http.ListenAndServe(addr, mux))

}
