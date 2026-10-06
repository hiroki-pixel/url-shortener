package handler

import (
	"errors"
	"net/http"

	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/hiroki-pixel/url-shortener/service"
)

// GET {short_code}のハンドラー関数の設定
func HandleRedirect(w http.ResponseWriter, r *http.Request, queries *database.Queries) {
	//リクエストから、パスパラメータに設定しているshort_codeの値を取得
	shortCode := r.PathValue("short_code")

	input := service.RedirectInput{
		ShortCode: shortCode,
	}

	output, err := service.ResolveRedirect(r.Context(), queries, input)

	if errors.Is(err, service.ErrShortCodeRequired) {
		http.Error(w, "shortcodeは必須項目です", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrShortCodeLength) {
		http.Error(w, "shortcodeは6桁固定です", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrShortCodeForm) {
		http.Error(w, "shortcodeは英数字のみにより構成されます", http.StatusBadRequest)
		return
	}

	if errors.Is(err, service.ErrShortCodeNotFound) {
		http.Error(w, "該当するshortcodeは登録されていません", http.StatusNotFound)
		return
	}

	if errors.Is(err, service.ErrDB) {
		http.Error(w, "サーバ内部でエラーが発生しました。時間を置いて再度試してください。", http.StatusInternalServerError)
		return
	}

	if err != nil {
		//その他のエラー
		http.Error(w, "予期せぬサーバエラーが発生しました。時間をおいて再度試してください", http.StatusInternalServerError)
		return
	}

	//302の返却およびLocationヘッダーにoriginalURLを設定してリダイレクトの実施
	http.Redirect(w, r, output.OriginalURL, http.StatusFound)

}
