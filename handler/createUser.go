package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/service"
)

// POST usersハンドラー関数の設定
func HandleCreateUser(w http.ResponseWriter, r *http.Request, userService *service.UserService) {
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

	//serviceに、リクエスト内容をインプット情報として渡す
	input := service.CreateUserInput{
		Username: request.Username,
		Password: request.Password,
	}

	//serviceを呼び出して、アウトプット情報を受け取る
	output, err := userService.CreateUser(r.Context(), input)

	//serviceからエラー返却があった場合に、handlerにてどのようなレスポンスを実施するか振り分ける
	if errors.Is(err, service.ErrUsernameRequired) {
		http.Error(w, "ユーザ名とパスワードは両方必須項目です", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrUsernameLength) {
		http.Error(w, "ユーザ名は5桁以上15桁以下としてください", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrPasswordRequired) {
		http.Error(w, "ユーザ名とパスワードは両方必須項目です", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrPasswordLength) {
		http.Error(w, "パスワードは8桁以上64桁以下としてください", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrPasswordForm) {
		http.Error(w, "パスワードは英数字を両方含むように設定してください", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrPasswordFailedToHash) {
		http.Error(w, "サーバ内部の処理エラーが発生しました。時間を置いて再度試してください。", http.StatusInternalServerError)
		return
	}

	if errors.Is(err, service.ErrUsernameConflict) {
		http.Error(w, "すでにユーザ名が使用されています。別のユーザ名を設定してください", http.StatusConflict)
		return
	}

	if errors.Is(err, service.ErrDB) {
		http.Error(w, "サーバ内部の処理エラーが発生しました。時間を置いて再度試してください。", http.StatusInternalServerError)
		return
	}
	//想定外のエラーが発生した際に吸収するようのエラーハンドリング
	if err != nil {
		http.Error(w, "サーバ内部の予期せぬ処理エラーが発生しました。時間を置いて再度試してください。", http.StatusInternalServerError)
		return
	}

	//serviceから受け取ったレスポンス情報(output)をHandler側の構造体に格納する
	type createUserResponse struct {
		UserID    uuid.UUID `json:"user_id"`
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"created_at"`
	}

	response := createUserResponse{
		UserID:    output.UserID,
		Username:  output.Username,
		CreatedAt: output.CreatedAt,
	}

	//レスポンス用にエンコードして、返却する
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to Encode Response %v\n", err)
	}
}
