package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/hiroki-pixel/url-shortener/internal/requestcontext"
)

// GET urlsのハンドラー関数の設定
func HandleListURL(w http.ResponseWriter, r *http.Request, queries *database.Queries, baseURL string) {
	//認証済みリクエストのContextからuser_idを取り出す。
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "認証情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	//ListsURLメソッドの実施（userIDに紐づくurl等情報の取得）
	urlRows, err := queries.ListsURLByUserID(r.Context(), userID)

	//DBへの検索処理についてerrのハンドリングをする
	//検索した結果何もない、というのは正常系

	if err != nil {
		//その他のエラー
		http.Error(w, "サーバエラーが発生しました。しばらく時間をおいて再度試してください", http.StatusInternalServerError)
		return
	}

	//レスポンスのための構造体を作成
	type URLResponse struct {
		ShortCode   string `json:"short_code"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		Memo        string `json:"memo"`
	}

	//今回のレスポンスは複数件返す可能性があるため、構造体を詰め込むスライスを作成する必要がある
	//1件もappendされなかったときに空スライス[]を返せるように、初期化しておく。
	urlResponses := []URLResponse{}

	//レスポンス情報のために、short_urlを作成
	//BASE_URL + short_code を行い、short_urlを作成しつつ、レスポンス情報を作成する
	for _, row := range urlRows {
		shortURL, err := url.JoinPath(baseURL, row.ShortCode)
		if err != nil {
			http.Error(w, "レスポンス情報の生成に失敗しました。再度時間をおいて試してください", http.StatusInternalServerError)
			return
		}

		//MemoをDBに格納しているsql.NullString型から、API用のstring型に変換
		memo := ""
		if row.Memo.Valid {
			memo = row.Memo.String
		}

		urlResponse := URLResponse{
			ShortCode:   row.ShortCode,
			ShortURL:    shortURL,
			OriginalURL: row.OriginalUrl,
			Memo:        memo,
		}
		urlResponses = append(urlResponses, urlResponse)
	}

	//レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urlResponses); err != nil {
		log.Println("レスポンス生成に失敗:", err)
	}

}
