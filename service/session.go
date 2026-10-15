package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/hiroki-pixel/url-shortener/repository"
	"golang.org/x/crypto/bcrypt"
)

// serviceがrepositoryを持つ状態にする
type SessionService struct {
	repo *repository.SessionRepository
}

// serviceを作成する関数を定義する
func NewSessionService(repo *repository.SessionRepository) *SessionService {
	return &SessionService{
		repo: repo,
	}
}

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

func (s *SessionService) CreateSession(ctx context.Context, input CreateSessionInput) (CreateSessionOutput, error) {
	//セッション作成の中身を定義
	//usernameのバリデーションチェック
	if input.Username == "" {
		return CreateSessionOutput{}, ErrUsernameRequired
	}

	//passwordのバリデーションチェック
	if input.Password == "" {
		return CreateSessionOutput{}, ErrPasswordRequired
	}

	//repositoryのsession.goのFindUserByUsernameメソッドを利用して該当のusername/passwordの組み合わせが保存済みか検索
	findUserInput := repository.FindUserByUsernameInput{
		Username: input.Username,
	}
	findUserOutput, err := s.repo.FindUserByUsername(ctx, findUserInput)
	if errors.Is(err, repository.ErrUserNotFound) {
		return CreateSessionOutput{}, ErrInvalidCredentials
	}

	if err != nil {
		return CreateSessionOutput{}, ErrDB
	}

	//検索で得た、保存済みpassword_hashとリクエストで受けとったpasswordを照合する
	err = bcrypt.CompareHashAndPassword([]byte(findUserOutput.PasswordHash), []byte(input.Password))
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

	//repositoryのsession.goのCreateSessionメソッドを使って、sessionsテーブルへの格納処理
	createSessionInput := repository.CreateSessionInput{
		SessionID: sessionID,
		UserID:    findUserOutput.UserID,
	}

	err = s.repo.CreateSession(ctx, createSessionInput)
	if errors.Is(err, repository.ErrDB) {
		return CreateSessionOutput{}, ErrDB
	}

	if err != nil {
		return CreateSessionOutput{}, ErrDB
	}

	//すべてのチェックが通った後、CreateSessionOutputに情報を詰める
	output := CreateSessionOutput{
		SessionID: sessionID,
	}

	return output, nil
}
