package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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
func HandleCreateUser(w http.ResponseWriter, r *http.Request, queries *database.Queries) {
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

	//user_idを払い出す。また、パスワードをハッシュ化する。
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
