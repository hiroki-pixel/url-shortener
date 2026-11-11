package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/repository"
	"golang.org/x/crypto/bcrypt"
)

// serviceがrepositoryを持つ状態にする
type SessionService struct {
	repo SessionRepositoryInterface
}

// 具体のSessionRepositoryに依存しないinterfaceとして定義する
type SessionRepositoryInterface interface {
	GetSessionByID(
		ctx context.Context,
		input repository.GetSessionByIDInput,
	) (repository.GetSessionByIDOutput, error)

	FindUserByUsername(
		ctx context.Context,
		input repository.FindUserByUsernameInput,
	) (repository.FindUserByUsernameOutput, error)

	CreateSession(
		ctx context.Context,
		input repository.CreateSessionInput,
	) error
}

// テスト用にFake Repositoryを作成する
type FakeSessionRepository struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
	ReturnErr error
}

func (f *FakeSessionRepository) GetSessionByID(
	ctx context.Context,
	input repository.GetSessionByIDInput,
) (repository.GetSessionByIDOutput, error) {

	// テスト用の戻り値を記載する
	return repository.GetSessionByIDOutput{
		UserID:    f.UserID,
		ExpiresAt: f.ExpiresAt,
	}, f.ReturnErr
}

func (f *FakeSessionRepository) CreateSession(
	ctx context.Context,
	input repository.CreateSessionInput,
) error {
	return nil
}

func (f *FakeSessionRepository) FindUserByUsername(
	ctx context.Context,
	input repository.FindUserByUsernameInput,
) (repository.FindUserByUsernameOutput, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	if err != nil {
		return repository.FindUserByUsernameOutput{}, err
	}
	return repository.FindUserByUsernameOutput{
		UserID:       uuid.New(),
		PasswordHash: string(hash),
	}, nil
}

// serviceを作成する関数(コンストラクタ)を定義する
func NewSessionService(repo SessionRepositoryInterface) *SessionService {
	return &SessionService{
		repo: repo,
	}
}

// 引数を構造体としてひとまとまりに定義する。(セッションの作成)
type CreateSessionInput struct {
	Username string
	Password string
}

// 戻り値の構造体のフィールドはHandlerに渡したいsession_idのみ(セッションの作成)
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

// 引数を構造体として定義する(セッションの有効期限確認)
type AuthenticateSessionInput struct {
	SessionID string
}

// mainに渡したい返り値を構造体として定義する(セッションの有効期限確認)
type AuthenticateSessionOutput struct {
	UserID uuid.UUID
}

// sessionが存在しなかった際のエラー
var ErrSessionNotFound = errors.New("session not found")

// sessionの有効期限切れが起こってた際のエラー
var ErrSessionExpired = errors.New("session is invalid")

func (s *SessionService) AuthenticateSession(ctx context.Context, input AuthenticateSessionInput) (AuthenticateSessionOutput, error) {
	//repositoryのsession.goのGetSessionByIDメソッドを呼び出し
	authenticateSessionInput := repository.GetSessionByIDInput{
		SessionID: input.SessionID,
	}
	authenticateSessionOutput, err := s.repo.GetSessionByID(ctx, authenticateSessionInput)
	if errors.Is(err, repository.ErrSessionNotFound) {
		return AuthenticateSessionOutput{}, ErrSessionNotFound
	}

	if err != nil {
		return AuthenticateSessionOutput{}, ErrDB
	}

	if time.Now().After(authenticateSessionOutput.ExpiresAt) {
		return AuthenticateSessionOutput{}, ErrSessionExpired
	}

	output := AuthenticateSessionOutput{
		UserID: authenticateSessionOutput.UserID,
	}

	return output, nil

}
