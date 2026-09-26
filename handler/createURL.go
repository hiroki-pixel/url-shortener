package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/hiroki-pixel/url-shortener/internal/requestcontext"
	"github.com/lib/pq"
)

// POST urlsのハンドラー関数の設定。
// short_urlの生成にBASE_URLが必要なので、main関数でbaseURLとして取得確認したものも引数にする
func HandleCreateURL(w http.ResponseWriter, r *http.Request, queries *database.Queries, baseURL string) {
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

	//memoは、未入力でリクエストされた場合、NULLとして扱うようにする。
	//入力値についてのAP側の意味やルールによる判断のため、DB側でNULLIF""を使うと暗黙知になると判断
	memo := sql.NullString{
		String: request.Memo,
		Valid:  request.Memo != "",
	}

	//入力内容のバリデーションを行う
	//memoについて
	if len(request.Memo) > 256 {
		http.Error(w, "memoは256桁以内で登録してください", http.StatusBadRequest)
		return
	}

	//original_urlについて
	if request.OriginalURL == "" {
		http.Error(w, "original_urlは必須項目です", http.StatusBadRequest)
		return
	}

	if len(request.OriginalURL) > 2048 {
		http.Error(w, "original_urlが登録可能な長さを超過しています", http.StatusBadRequest)
		return
	}

	//URLの形式が正しいかについてのチェック
	//parsedURLへの変換とスキーマ、ホスト部の確認
	parsedURL, err := url.Parse(request.OriginalURL)
	if err != nil {
		http.Error(w, "URLとして認識されませんでした、入力内容を確認してください", http.StatusBadRequest)
		return
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		http.Error(w, "スキームはhttpもしくはhttpsとしてください、入力内容を確認してください", http.StatusBadRequest)
		return
	}

	if parsedURL.Host == "" {
		http.Error(w, "host部分が確認できませんでした、入力内容を確認してください", http.StatusBadRequest)
		return
	}

	//url_idを払い出す。
	urlID := uuid.New()

	//short_codeの払出し処理。
	//ランダムなn桁な英数字の文字列の生成関数randAlphaNumeric関数
	//使用許可設定した英数字からランダムに選んで一つずつ詰めていく作業をしている
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	randAlphaNumeric := func(n int) (string, error) {
		result := make([]byte, n)

		for i := 0; i < n; i++ {
			randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
			if err != nil {
				return "", err
			}

			result[i] = letters[randomIndex.Int64()]
		}

		return string(result), nil
	}

	//randAlphaNumeric関数を使用してshort_codeを生成
	shortCode, err := randAlphaNumeric(6)
	if err != nil {
		http.Error(w, "短縮コードの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	//short_urlの生成
	//BASE_URL + short_code を行い、short_urlを作成する
	shortURL, err := url.JoinPath(baseURL, shortCode)
	if err != nil {
		log.Println("shortURL生成に失敗")
		http.Error(w, "短縮URLの生成に失敗しました。再度試してください", http.StatusInternalServerError)
		return
	}

	//InsertURLメソッドを利用して、urlsテーブルにInsert
	//引数となる、構造体を用意
	insertParams := database.InsertURLParams{
		UrlID:       urlID,
		UserID:      userID,
		OriginalUrl: request.OriginalURL,
		Memo:        memo,
		ShortCode:   shortCode,
	}
	err = queries.InsertURL(r.Context(), insertParams)

	//Insert時にエラー発生した際のエラーハンドリング
	//user_id + original_url でのUNIQUE制約に引っかかった場合（Conflict）
	// sqlcのInsertURLから返されたDBエラーを判定する
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Constraint == "urls_user_id_original_url_key" {
		log.Println("You registered original_url before. Failed to INSERT URLRow ")
		http.Error(w, "URLが既に登録されています。重複したURLの登録は禁止されています。", http.StatusConflict)
		return
	}

	//その他のエラー(InternalServerError)
	if err != nil {
		log.Println("Failed to INSERT URLRow")
		http.Error(w, "サーバ内部のエラーが発生しました。時間をおいて再度試してください", http.StatusInternalServerError)
		return
	}

	//responseの返却
	//responseの構造体を作成
	type registerURLResponse struct {
		ShortCode string `json:"short_code"`
		ShortURL  string `json:"short_url"`
	}
	response := registerURLResponse{
		ShortCode: shortCode,
		ShortURL:  shortURL,
	}

	//responseに内容を詰め込む
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Failed to Encode Response %v\n", err)
	}

}
