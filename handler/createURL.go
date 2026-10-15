package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/hiroki-pixel/url-shortener/internal/requestcontext"
	"github.com/hiroki-pixel/url-shortener/service"
)

// POST urlsのハンドラー関数の設定。
// short_urlの生成にBASE_URLが必要なので、main関数でbaseURLとして取得確認したものも引数にする
func HandleCreateURL(w http.ResponseWriter, r *http.Request, urlService *service.URLService, baseURL string) {
	// 認証済みリクエストのコンテキストに登録されているuserIDを取り出す
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "認証情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	//リクエスト内容をGoで扱えるようにする
	type createURLRequest struct {
		OriginalURL string `json:"original_url"`
		Memo        string `json:"memo"`
	}
	request := createURLRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	//serviceを呼び出すための、インプット情報を作る
	input := service.CreateURLInput{
		UserID:      userID,
		OriginalURL: request.OriginalURL,
		Memo:        request.Memo,
		BaseURL:     baseURL,
	}

	//serviceを呼び出して、アウトプット情報を受け取る
	output, err := urlService.CreateURL(r.Context(), input)

	//serviceからエラー返却があった場合に、handlerにてどのようなレスポンスを実施するか振り分ける
	if errors.Is(err, service.ErrOriginalURLRequired) {
		http.Error(w, "originalURLは必須項目です", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrOriginalURLLength) {
		http.Error(w, "originalURLは2048桁以内のものを登録してください", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrMemoLength) {
		http.Error(w, "memoは256桁以内のものを登録してください", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrURLParse) {
		http.Error(w, "URL形式として正しくありません。確認して再度実行してください", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrURLScheme) {
		http.Error(w, "URLのスキーマがhttpもしくはhttpsであることを確認して再度実行してください", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrURLHost) {
		http.Error(w, "URLのホストが入力されていません。入力URLを確認して再度実行してください", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrGenerateShortCode) {
		http.Error(w, "短縮コードの生成に失敗しました。時間を置いて再度試してください", http.StatusInternalServerError)
		return
	}

	if errors.Is(err, service.ErrGenerateShortURL) {
		http.Error(w, "短縮URLの生成に失敗しました。時間を置いて再度試してください", http.StatusInternalServerError)
		return
	}

	if errors.Is(err, service.ErrOriginalURLDuplicate) {
		http.Error(w, "入力URLはすでに登録済みです", http.StatusConflict)
		return
	}

	if errors.Is(err, service.ErrDB) {
		http.Error(w, "サーバ内部でエラーが発生しました。時間を置いて再度試してください", http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, "想定外エラーが発生しました。時間を置いて再度試してください", http.StatusInternalServerError)
		return
	}

	//responseの返却
	//responseの構造体を作成
	type registerURLResponse struct {
		ShortCode string `json:"short_code"`
		ShortURL  string `json:"short_url"`
	}
	response := registerURLResponse{
		ShortCode: output.ShortCode,
		ShortURL:  output.ShortURL,
	}

	//responseに内容を詰め込む
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to Encode Response %v\n", err)
	}

}
