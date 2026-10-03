package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"net/url"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/lib/pq"
)

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
