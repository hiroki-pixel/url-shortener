package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"

	"github.com/hiroki-pixel/url-shortener/internal/database"
)

// GET {short_code}のハンドラー関数の設定
func HandleRedirect(w http.ResponseWriter, r *http.Request, queries *database.Queries) {
	//リクエストから、パスパラメータに設定しているshort_codeの値を取得
	shortCode := r.PathValue("short_code")
	//shortCodeの値についてバリデーションチェックをする
	if shortCode == "" {
		http.Error(w, "short_codeを入力してください", http.StatusBadRequest)
		return
	}
	if len(shortCode) != 6 {
		http.Error(w, "short_codeは6桁です、再度確認して入力してください", http.StatusBadRequest)
		return
	}
	matched, err := regexp.MatchString(`^[a-zA-Z0-9]+$`, shortCode)
	if err != nil {
		http.Error(w, "サーバエラーが発生しました。時間をおいて試してください", http.StatusInternalServerError)
		return
	}

	if !matched {
		http.Error(w, "short_codeは英数字で入力してください", http.StatusBadRequest)
		return
	}

	//sqlc生成メソッドを使用して、original_urlを取得
	originalURL, err := queries.GetOriginURLByShortCode(r.Context(), shortCode)
	//DBへの検索処理についてerrのハンドリングをする
	// GetOriginURLByShortCodeから返されたDBエラーを判定する
	if errors.Is(err, sql.ErrNoRows) {
		//リクエストで受けとったshort_codeでの検索結果が存在しなかった
		http.Error(w, "無効なshort_codeです", http.StatusNotFound)
		return
	}

	if err != nil {
		//その他のエラー
		http.Error(w, "サーバエラーが発生しました。時間をおいて再度試してください", http.StatusInternalServerError)
		return
	}

	//302の返却およびLocationヘッダーにoriginalURLを設定してリダイレクトの実施
	http.Redirect(w, r, originalURL, http.StatusFound)

}
