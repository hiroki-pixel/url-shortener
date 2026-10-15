package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/hiroki-pixel/url-shortener/internal/requestcontext"
	"github.com/hiroki-pixel/url-shortener/service"
)

// GET urlsのハンドラー関数の設定
func HandleListURL(w http.ResponseWriter, r *http.Request, urlService *service.URLService, baseURL string) {
	//認証済みリクエストのContextからuser_idを取り出す。
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "認証情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	input := service.ListURLInput{
		UserID:  userID,
		BaseURL: baseURL,
	}

	output, err := urlService.ListURL(r.Context(), input)

	if errors.Is(err, service.ErrDB) {
		http.Error(w, "サーバ内部でエラーが発生しました。時間を置いて再度試してください。", http.StatusInternalServerError)
		return
	}

	if errors.Is(err, service.ErrGenerateShortURL) {
		http.Error(w, "サーバ内部でエラーが発生しました。時間を置いて再度試してください。", http.StatusInternalServerError)
		return
	}

	//レスポンスのための構造体を作成
	type listURLResponse struct {
		ShortCode   string `json:"short_code"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		Memo        string `json:"memo"`
	}

	responses := []listURLResponse{}
	for _, item := range output.URLs {
		response := listURLResponse{
			ShortCode:   item.ShortCode,
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
			Memo:        item.Memo,
		}

		responses = append(responses, response)
	}

	//レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		log.Println("レスポンス生成に失敗:", err)
	}

}
