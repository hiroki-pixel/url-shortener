package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"net/url"
	"regexp"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/lib/pq"
)

//createURLについて

// handlerから受け取る情報を構造体として定義する
type CreateURLInput struct {
	UserID      uuid.UUID
	OriginalURL string
	Memo        string
	BaseURL     string
}

// handlerに返却する情報を構造体として定義する
type CreateURLOutput struct {
	ShortCode string
	ShortURL  string
}

// エラーを定義する
// originalURLの必須エラーは共通errorsで実装
// memoやoriginalurlの桁数オーバーについてのエラー
var ErrMemoLength = errors.New("memo length is invalid")
var ErrOriginalURLLength = errors.New("original url length is invalid")
var ErrOriginalURLDuplicate = errors.New("original url is duplicated")
var ErrURLParse = errors.New("URL parse failed")
var ErrURLScheme = errors.New("URL Scheme is invalid")
var ErrURLHost = errors.New("URL Host is missing")
var ErrGenerateShortCode = errors.New("failed to generate short code")
var ErrGenerateShortURL = errors.New("failed to generate short URL")

func CreateURL(ctx context.Context, queries *database.Queries, input CreateURLInput) (CreateURLOutput, error) {
	//OriginalURLからshortcodeの作成およびその保持処理についての中身を定義
	//originalurlのバリデーションチェック
	if input.OriginalURL == "" {
		return CreateURLOutput{}, ErrOriginalURLRequired
	}
	if len(input.OriginalURL) > 2048 {
		return CreateURLOutput{}, ErrOriginalURLLength
	}

	//memoについてのバリデーションチェック
	if len(input.Memo) > 256 {
		return CreateURLOutput{}, ErrMemoLength
	}

	//URLの形式が正しいかについてのチェック
	//parsedURLへの変換とスキーマ、ホスト部の確認
	parsedURL, err := url.Parse(input.OriginalURL)
	if err != nil {
		return CreateURLOutput{}, ErrURLParse
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return CreateURLOutput{}, ErrURLScheme
	}

	if parsedURL.Host == "" {
		return CreateURLOutput{}, ErrURLHost
	}

	//memoは、未入力でリクエストされた場合、NULLとして扱うようにする。
	//入力値についてのAP側の意味やルールによる判断のため、DB側でNULLIF""を使うと暗黙知になると判断
	memo := sql.NullString{
		String: input.Memo,
		Valid:  input.Memo != "",
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
		return CreateURLOutput{}, ErrGenerateShortCode
	}

	//short_urlの生成
	//BASE_URL + short_code を行い、short_urlを作成する
	shortURL, err := url.JoinPath(input.BaseURL, shortCode)
	if err != nil {
		return CreateURLOutput{}, ErrGenerateShortURL
	}

	//InsertURLメソッドを利用して、urlsテーブルにInsert
	insertParams := database.InsertURLParams{
		UrlID:       urlID,
		UserID:      input.UserID,
		OriginalUrl: input.OriginalURL,
		Memo:        memo,
		ShortCode:   shortCode,
	}

	err = queries.InsertURL(ctx, insertParams)
	//Insert時にエラー発生した際のエラーハンドリング
	//user_id + original_url でのUNIQUE制約に引っかかった場合（Conflict）
	// sqlcのInsertURLから返されたDBエラーを判定する
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Constraint == "urls_user_id_original_url_key" {
		return CreateURLOutput{}, ErrOriginalURLDuplicate
	}

	//その他のエラー(InternalServerError)
	if err != nil {
		return CreateURLOutput{}, ErrDB
	}

	//すべてのチェックが通った後、CreateURLOutputに情報を詰める
	output := CreateURLOutput{
		ShortCode: shortCode,
		ShortURL:  shortURL,
	}

	return output, nil
}

// ListURLについて
// handlerから受け取る情報を構造体として定義する
type ListURLInput struct {
	UserID  uuid.UUID
	BaseURL string
}

// handlerへ返却する情報を構造体として定義する(出力が複数件になるのでまとめる構造体も作成する)
type ListURLItem struct {
	ShortCode   string
	ShortURL    string
	OriginalURL string
	Memo        string
}

type ListURLOutput struct {
	URLs []ListURLItem
}

//エラーの定義
//DB接続の汎用的なエラーは共通エラーを使用する
//shortURLの作成エラーはcreateURLと共通のものを使用する

// 関数内で定義する変数名の意味。item = 検索結果の1行、items = 検索結果すべて（複数行）、output = 検索結果すべてを返すための箱
func ListURL(ctx context.Context, queries *database.Queries, input ListURLInput) (ListURLOutput, error) {
	//ListsURLメソッドの実施（userIDに紐づくurl等情報の取得）
	urlRows, err := queries.ListsURLByUserID(ctx, input.UserID)
	if err != nil {
		return ListURLOutput{}, ErrDB
	}

	//検索結果のURL情報を詰めるためのスライスの作成
	items := []ListURLItem{}

	//レスポンス情報のために、short_urlを作成
	//BASE_URL + short_code を行い、short_urlを作成しつつ、レスポンス情報を作成する
	for _, row := range urlRows {
		shortURL, err := url.JoinPath(input.BaseURL, row.ShortCode)
		if err != nil {
			return ListURLOutput{}, ErrGenerateShortURL
		}

		//MemoをDBに格納しているsql.NullString型から、API用のstring型に変換
		memo := ""
		if row.Memo.Valid {
			memo = row.Memo.String
		}

		item := ListURLItem{
			ShortCode:   row.ShortCode,
			ShortURL:    shortURL,
			OriginalURL: row.OriginalUrl,
			Memo:        memo,
		}

		items = append(items, item)
	}
	output := ListURLOutput{
		URLs: items,
	}

	return output, nil

}

// redirectについて
// handlerから受け取る情報を構造体として定義する
type RedirectInput struct {
	ShortCode string
}

// handlerへ返却する情報を構造体として定義する
type RedirectOutput struct {
	OriginalURL string
}

// エラーについての定義
// shortcodeが存在しないエラーについては、共通のerrorsにて定義
var ErrShortCodeLength = errors.New("shortcode length is invalid")
var ErrShortCodeForm = errors.New("shortcode form is invalid")
var ErrShortCodeNotFound = errors.New("shortcode is not found")

func ResolveRedirect(ctx context.Context, queries *database.Queries, input RedirectInput) (RedirectOutput, error) {
	if input.ShortCode == "" {
		return RedirectOutput{}, ErrShortCodeRequired
	}

	if len(input.ShortCode) != 6 {
		return RedirectOutput{}, ErrShortCodeLength
	}

	matched, err := regexp.MatchString(`^[a-zA-Z0-9]+$`, input.ShortCode)
	if err != nil {
		return RedirectOutput{}, err
	}
	if !matched {
		return RedirectOutput{}, ErrShortCodeForm
	}

	//sqlc生成メソッドを使用して、original_urlを取得
	originalURL, err := queries.GetOriginURLByShortCode(ctx, input.ShortCode)
	//DBへの検索処理についてerrのハンドリングをする
	// GetOriginURLByShortCodeから返されたDBエラーを判定する
	if errors.Is(err, sql.ErrNoRows) {
		return RedirectOutput{}, ErrShortCodeNotFound
	}

	if err != nil {
		return RedirectOutput{}, ErrDB
	}

	output := RedirectOutput{
		OriginalURL: originalURL,
	}
	return output, nil
}
