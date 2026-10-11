package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/lib/pq"
)

//serviceから受け取る情報をひとまとまりに定義する。（createURL）

type CreateURLInput struct {
	UrlID       uuid.UUID
	UserID      uuid.UUID
	OriginalURL string
	Memo        string
	ShortCode   string
}

//serviceへ返す情報をひとまとまりに定義するが、DBから帰ってくる値はないので定義不要。（createURL）

// 発生する可能性のあるエラーを定義（createURL）
var ErrOriginalURLDuplicate = errors.New("original url is duplicated")

func CreateURL(ctx context.Context, queries *database.Queries, input CreateURLInput) error {
	memo := sql.NullString{
		String: input.Memo,
		Valid:  input.Memo != "",
	}

	insertParams := database.InsertURLParams{
		UrlID:       input.UrlID,
		UserID:      input.UserID,
		OriginalUrl: input.OriginalURL,
		Memo:        memo,
		ShortCode:   input.ShortCode,
	}

	err := queries.InsertURL(ctx, insertParams)
	//Insert時にエラー発生した際のエラーハンドリング
	//user_id + original_url でのUNIQUE制約に引っかかった場合（Conflict）
	// sqlcのInsertURLから返されたDBエラーを判定する
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Constraint == "urls_user_id_original_url_key" {
		return ErrOriginalURLDuplicate
	}

	//その他のエラー(InternalServerError)
	if err != nil {
		return ErrDB
	}
	return nil
}

// serviceから受け取る情報をひとまとまりに定義する。（ListsURLByUserID）
type ListURLByUserIDInput struct {
	UserID uuid.UUID
}

// serviceへ返す情報をひとまとまりに定義する。（ListsURLByUserID）
type ListURLItem struct {
	ShortCode   string
	OriginalURL string
	Memo        string
}

type ListURLByUserIDOutput struct {
	URLs []ListURLItem
}

func ListURLByUserID(ctx context.Context, queries *database.Queries, input ListURLByUserIDInput) (ListURLByUserIDOutput, error) {
	urlRows, err := queries.ListsURLByUserID(ctx, input.UserID)
	if err != nil {
		return ListURLByUserIDOutput{}, ErrDB
	}

	items := []ListURLItem{}

	for _, row := range urlRows {
		memo := ""
		if row.Memo.Valid {
			memo = row.Memo.String
		}

		item := ListURLItem{
			ShortCode:   row.ShortCode,
			OriginalURL: row.OriginalUrl,
			Memo:        memo,
		}

		items = append(items, item)
	}
	output := ListURLByUserIDOutput{
		URLs: items,
	}

	return output, nil
}

// serviceから受け取る情報をひとまとまりに定義する。（GetOriginURLByShortCode）
type GetOriginURLByShortCodeInput struct {
	ShortCode string
}

// serviceへ返す情報をひとまとまりに定義する。（GetOriginURLByShortCode）
type GetOriginURLByShortCodeOutput struct {
	OriginalURL string
}

// 発生する可能性のあるエラーを定義する。（GetOriginURLByShortCode）
var ErrOriginalURLNotFound = errors.New("original url is not found")

func GetOriginURLByShortCode(ctx context.Context, queries *database.Queries, input GetOriginURLByShortCodeInput) (GetOriginURLByShortCodeOutput, error) {
	originalURL, err := queries.GetOriginURLByShortCode(ctx, input.ShortCode)

	if errors.Is(err, sql.ErrNoRows) {
		return GetOriginURLByShortCodeOutput{}, ErrOriginalURLNotFound
	}
	if err != nil {
		return GetOriginURLByShortCodeOutput{}, ErrDB
	}

	output := GetOriginURLByShortCodeOutput{
		OriginalURL: originalURL,
	}

	return output, nil

}
