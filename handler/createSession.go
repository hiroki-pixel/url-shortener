package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/hiroki-pixel/url-shortener/service"
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

	//serviceのCreateSessionに渡すためのinputを作成する
	input := service.CreateSessionInput{
		Username: request.Username,
		Password: request.Password,
	}

	//serviceのCreateSessionを呼び出す
	output, err := service.CreateSession(r.Context(), queries, input)

	//serviceからエラー返却があった場合のハンドリングを行う
	if errors.Is(err, service.ErrUsernameRequired) {
		http.Error(w, "ユーザ名とパスワードは両方必須項目です", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrPasswordRequired) {
		http.Error(w, "ユーザ名とパスワードは両方必須項目です", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrInvalidCredentials) {
		http.Error(w, "ユーザ名もしくはパスワードが無効です", http.StatusUnauthorized)
		return
	}

	if errors.Is(err, service.ErrSessionIDGeneration) {
		http.Error(w, "サーバ内部でエラーが発生しました。時間を置いて再度試してください", http.StatusInternalServerError)
		return
	}

	if errors.Is(err, service.ErrDB) {
		http.Error(w, "サーバ内部でエラーが発生しました。時間を置いて再度試してください", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "サーバ内部の予期せぬ処理エラーが発生しました。時間を置いて再度試してください。", http.StatusInternalServerError)
		return
	}

	//session_idの情報をCookieに入れる。また、cookieの設定をする
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: output.SessionID,
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
