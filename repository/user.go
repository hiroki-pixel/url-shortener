package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/lib/pq"
)

// serviceから受け取る情報をひとまとまりに定義する。
type CreateUserInput struct {
	Username     string
	UserID       uuid.UUID
	PasswordHash string
}

// serviceへ返す情報をひとまとまりに定義する。
type CreateUserOutput struct {
	CreatedAt time.Time
}

// 返却する可能性のあるエラーを定義する。
// 汎用的なDBエラーについては、repository共通のerrorsにて実装
var ErrUsernameConflict = errors.New("username is conflict")

func CreateUser(ctx context.Context, queries *database.Queries, input CreateUserInput) (CreateUserOutput, error) {

	//DBへのユーザ情報格納、CreatedAtの取得
	insertParams := database.InsertUserParams{
		UserID:       input.UserID,
		Username:     input.Username,
		PasswordHash: input.PasswordHash,
	}

	createdAt, err := queries.InsertUser(ctx, insertParams)

	//InsertUserから返されたDBエラーのハンドリングをする
	var pqErr *pq.Error

	if errors.As(err, &pqErr) && pqErr.Constraint == "users_username_key" {
		log.Println("Your username is not Unique. Failed to INSERT UserRow ")
		return CreateUserOutput{}, ErrUsernameConflict
	}

	if err != nil {
		log.Println("Failed to INSERT UserRow")
		return CreateUserOutput{}, ErrDB
	}

	output := CreateUserOutput{
		CreatedAt: createdAt,
	}

	return output, nil

}
