package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
)

// serviceから受け取る情報をひとまとまりに定義する。(ユーザネームによる検索)
type FindUserByUsernameInput struct {
	Username string
}

// serviceへ返す情報をひとまとまりに定義する。(ユーザネームによる検索)
type FindUserByUsernameOutput struct {
	UserID       uuid.UUID
	PasswordHash string
}

// 発生しうるエラーを定義する（ユーザネームによる検索）
var ErrUserNotFound = errors.New("user not found")

func FindUserByUsername(ctx context.Context, queries *database.Queries, input FindUserByUsernameInput) (FindUserByUsernameOutput, error) {
	//username検索メソッドの利用
	getUserInfo, err := queries.GetUserByUsername(ctx, input.Username)
	if errors.Is(err, sql.ErrNoRows) {
		return FindUserByUsernameOutput{}, ErrUserNotFound
	}
	if err != nil {
		return FindUserByUsernameOutput{}, ErrDB
	}

	output := FindUserByUsernameOutput{
		UserID:       getUserInfo.UserID,
		PasswordHash: getUserInfo.PasswordHash,
	}
	return output, nil
}

// serviceから受け取る情報をひとまとまりに定義する。（セッション情報の格納）
type CreateSessionInput struct {
	SessionID string
	UserID    uuid.UUID
}

// serviceへ返す情報をひとまとまりに定義するが、DBから帰ってくる値はないので定義不要

func CreateSession(ctx context.Context, queries *database.Queries, input CreateSessionInput) error {
	//sessionsテーブルへの格納処理
	//InsertSessionメソッドが要求する構造体InsertSessionParams型の変数を作成。
	insertParams := database.InsertSessionParams{
		SessionID: input.SessionID,
		UserID:    input.UserID,
	}

	//sessionsテーブルへの格納処理実施
	err := queries.InsertSession(ctx, insertParams)
	if err != nil {
		return ErrDB
	}

	return nil
}
