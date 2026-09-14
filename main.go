package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

// 認証が必要な機能で使用する、認証+目的の機能に処理を飛ばす関数
func authMiddleware(queries *database.Queries, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//認証処理内容。ここが正常にいけば、目的の機能に飛ばすし、でなければ401エラーを返す。

		next.ServeHTTP(w, r)
	})
}

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
func handleCreateUser(w http.ResponseWriter, r *http.Request, queries *database.Queries) {
	//jsonで受けとったリクエスト情報をGoで扱えるようにDecodeする
	type createUserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	request := createUserRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Failed to Decode Request Body %v\n", err)
		http.Error(w, "不正なリクエストです\n", http.StatusBadRequest)
		return
	}
	//バリデーションチェックの実装
	//usernameのバリデーションチェック
	if request.Username == "" {
		http.Error(w, "usernameは必須項目です\n", http.StatusBadRequest)
		return
	}

	if len(request.Username) < 5 || len(request.Username) > 15 {
		http.Error(w, "usernameは5桁以上15桁以下で設定する必要があります\n", http.StatusBadRequest)
		return
	}

	//passwordのバリデーションチェック
	if request.Password == "" {
		http.Error(w, "passwordは必須項目です\n", http.StatusBadRequest)
		return
	}

	if len(request.Password) < 8 || len(request.Password) > 64 {
		http.Error(w, "passwordは8桁以上64桁以下で設定する必要があります\n", http.StatusBadRequest)
		return
	}

	if containsAlphaAndNumber(request.Password) == false {
		http.Error(w, "passwordは英数字を両方含む必要があります\n", http.StatusBadRequest)
		return
	}

	log.Printf("usernameは%sで、passwordは%sです\n", request.Username, "隠ぺい")

	//処理として、user_idを払い出す。また、パスワードをハッシュ化する。
	userID := uuid.New()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Failed to Create HashPassword")
		http.Error(w, "パスワードの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	//byte配列のハッシュパスワードを文字列に変換しておく
	passwordHashString := string(passwordHash)

	//sqlcの導入後のDBへの格納処理
	//sqlcが要求する引数になる構造体(InsertUserParams)の作成
	insertParams := database.InsertUserParams{
		UserID:       userID,
		Username:     request.Username,
		PasswordHash: passwordHashString,
	}
	//sqlcで作った、InsertUserメソッドを使う
	createdAt, err := queries.InsertUser(r.Context(), insertParams)

	// sqlcのInsertUserから返されたDBエラーを判定する
	var pqErr *pq.Error

	//usernameの重複エラーのエラーハンドリングをする
	if errors.As(err, &pqErr) && pqErr.Constraint == "users_username_key" {
		log.Println("Your username is not Unique. Failed to INSERT UserRow ")
		http.Error(w, "ユーザ名が既に使用されているためユーザ登録に失敗しました。別のユーザ名を使用してください", http.StatusConflict)
		return
	}
	//その他のINSERT分の中でエラーが発生した場合のエラーハンドリングをする。
	if err != nil {
		log.Println("Failed to INSERT UserRow")
		http.Error(w, "サーバエラーのためユーザ登録に失敗しました。時間をおいて再度試してください。", http.StatusInternalServerError)
		return
	}

	//レスポンス情報を作る
	type createUserResponse struct {
		UserID    uuid.UUID `json:"user_id"`
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"created_at"`
	}
	response := createUserResponse{}

	//レスポンス用の構造体のフィールドに、処理で作成した情報やリクエストで得た情報を詰め込む
	response.UserID = userID
	response.CreatedAt = createdAt
	response.Username = insertParams.Username

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to Encode Response %v\n", err)
	}
}

// POST sessionsハンドラー関数の設定
func handleCreateSession(w http.ResponseWriter, r *http.Request, queries *database.Queries) {
	type createSessionRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	//リクエスト内容をGoで扱えるようにする
	request := createSessionRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Failed to Decode Request Body %v\n", err)
		http.Error(w, "不正なリクエストです", http.StatusBadRequest)
		return
	}

	// POST sessionsでは必須項目の有無のみチェックする
	//ユーザ存在確認とパスワードの正当性の検証は後続処理で反映
	if request.Username == "" {
		log.Println("username is required")
		http.Error(w, "不正なリクエストです。ユーザネームは必須項目です", http.StatusBadRequest)
		return
	}

	if request.Password == "" {
		log.Println("password is required")
		http.Error(w, "不正なリクエストです。パスワードは必須項目です", http.StatusBadRequest)
		return
	}

	//usersテーブルへのusername検索およびpassword_hashの取得
	//sqlcで作成したusername検索メソッドの利用
	getUserInfo, err := queries.GetUserByUsername(r.Context(), request.Username)

	//DBへの検索処理についてerrのハンドリングをする
	// sqlcのGettUserByUsernameから返されたDBエラーを判定する
	if errors.Is(err, sql.ErrNoRows) {
		//リクエストで受けとったusernameが存在しなかった
		http.Error(w, "無効なユーザ名またはパスワードです", http.StatusUnauthorized)
		return
	}

	if err != nil {
		//その他のエラー
		http.Error(w, "サーバエラーが発生しました。しばらく時間をおいて再度試してください", http.StatusInternalServerError)
		return
	}

	//保存済みpassword_hashとリクエストで受けとったpasswordを照合する
	err = bcrypt.CompareHashAndPassword([]byte(getUserInfo.PasswordHash), []byte(request.Password))
	if err != nil {
		//リクエストで受けとったpasswordが無効なものだった
		http.Error(w, "無効なユーザ名またはパスワードです", http.StatusUnauthorized)
		return
	}

	//session_idの発行処理
	randomBytes := make([]byte, 32)

	_, err = rand.Read(randomBytes)
	if err != nil {
		http.Error(w, "セッションIDの払出に失敗しました。再度実施してください", http.StatusInternalServerError)
		return
	}
	//[]byte型から文字列に変換する
	sessionID := base64.RawURLEncoding.EncodeToString(randomBytes)

	//sessionsテーブルへの格納処理
	//InsertSessionメソッドが要求する構造体InsertSessionParams型の変数を作成。
	insertParams := database.InsertSessionParams{
		SessionID: sessionID,
		UserID:    getUserInfo.UserID,
	}

	//sessionsテーブルへの格納処理実施
	err = queries.InsertSession(r.Context(), insertParams)
	if err != nil {
		log.Println("DBへのsessionの格納失敗")
		http.Error(w, "サーバ内エラーです。しばらくしてもう一度試してください", http.StatusInternalServerError)
		return
	}

	//session_idの情報をCookieに入れる。また、cookieの設定をする
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: sessionID,
		/* Domain: ,　Domain部分は本番環境で設定する*/
		Path:     "/",
		MaxAge:   10800, //設定したセッションの持続時間
		HttpOnly: true,
		Secure:   false, //開発環境ではfalseだが、本番ではtrueにする必要あり
		SameSite: http.SameSiteLaxMode,
	}

	//レスポンスヘッダに設定したcookieを入れる
	http.SetCookie(w, cookie)

	//Post sessionsの成功レスポンスをする
	w.WriteHeader(http.StatusCreated)
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

	//DBとの接続窓口たるdbを使って、sqlcで作成したメソッドを持つqueriesにする。
	queries := database.New(db)

	//Routerの作成
	mux := http.NewServeMux()

	//作成したRouter(つまり、mux)にhandlerを登録する
	//Post /usersのルーティング
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		handleCreateUser(w, r, queries)
	})
	//Post /sessionsのルーティング
	mux.HandleFunc("POST /sessions", func(w http.ResponseWriter, r *http.Request) {
		handleCreateSession(w, r, queries)
	})

	//設定したアドレスとRouterを使用してサーバを起動
	addr := ":" + port
	log.Fatal(http.ListenAndServe(addr, mux))

}
