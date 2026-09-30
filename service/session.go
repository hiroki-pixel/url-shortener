package service

import (
	"context"
	"errors"

	"github.com/hiroki-pixel/url-shortener/internal/database"
)

// 引数を構造体としてひとまとまりに定義する。
type CreateSessionInput struct {
	Username string
	Password string
}

// 戻り値の構造体のフィールドはHandlerに渡したいsession_idのみ
type CreateSessionOutput struct {
	SessionID string
}

// 発生するエラーを定義する
// Username,Passwordの必須項目チェックは共通のerrorsに切り出し

// Username,Passwordが無効であった際のエラー
var ErrInvalidCredentials = errors.New("invalid credentials")

// session_idの払出に失敗した際のエラー
var ErrSessionIDGeneration = errors.New("session ID generation failed")

func CreateSession(ctx context.Context, queries *database.Queries, input CreateSessionInput) (CreateSessionOutput, error) {
	//セッション作成の中身を定義
}
