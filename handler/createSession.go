package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/hiroki-pixel/url-shortener/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// POST sessionsハンドラー関数の設定
func HandleCreateSession(w http.ResponseWriter, r *http.Request, queries *database.Queries) {
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
		/* Domain: ,　Domain部分は設定しない*/
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
