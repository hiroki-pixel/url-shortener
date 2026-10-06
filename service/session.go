package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"

	"github.com/hiroki-pixel/url-shortener/internal/database"
	"golang.org/x/crypto/bcrypt"
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
	//usernameのバリデーションチェック
	if input.Username == "" {
		return CreateSessionOutput{}, ErrUsernameRequired
	}

	//passwordのバリデーションチェック
	if input.Password == "" {
		return CreateSessionOutput{}, ErrPasswordRequired
	}

	//username検索メソッドの利用
	getUserInfo, err := queries.GetUserByUsername(ctx, input.Username)
	if errors.Is(err, sql.ErrNoRows) {
		return CreateSessionOutput{}, ErrInvalidCredentials
	}
	if err != nil {
		return CreateSessionOutput{}, ErrDB
	}

	//検索で得た、保存済みpassword_hashとリクエストで受けとったpasswordを照合する
	err = bcrypt.CompareHashAndPassword([]byte(getUserInfo.PasswordHash), []byte(input.Password))
	if err != nil {
		return CreateSessionOutput{}, ErrInvalidCredentials
	}

	//session_idの発行処理
	randomBytes := make([]byte, 32)

	_, err = rand.Read(randomBytes)
	if err != nil {
		return CreateSessionOutput{}, ErrSessionIDGeneration
	}
	//[]byte型から文字列に変換する
	sessionID := base64.RawURLEncoding.EncodeToString(randomBytes)

	//sessionsテーブルへの格納処理
	//InsertSessionメソッドが要求する構造体InsertSessionParams型の変数を作成。
	insertParams := database.InsertSessionParams{
		SessionID: sessionID,
		UserID:    getUserInfo.UserID,
	}

	//sessionsテーブルへの格納処理実施
	err = queries.InsertSession(ctx, insertParams)
	if err != nil {
		return CreateSessionOutput{}, ErrDB
	}

	//すべてのチェックが通った後、CreateSessionOutputに情報を詰める
	output := CreateSessionOutput{
		SessionID: sessionID,
	}

	return output, nil
}
