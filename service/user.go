package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/hiroki-pixel/url-shortener/repository"
	"golang.org/x/crypto/bcrypt"
)

// 引数を構造体としてひとまとまりに定義する。
type CreateUserInput struct {
	Username string
	Password string
}

// 戻り値を構造体としてひとまとまりに定義する。
type CreateUserOutput struct {
	UserID    uuid.UUID
	Username  string
	CreatedAt time.Time
}

// 発生するエラーを定義する。
//Username,Passwordの必須項目チェックは共通のerrorsに切り出し
// usernameに関するバリデーションチェックエラー

var ErrUsernameLength = errors.New("username length is invalid")

// passwordに関するバリデーションチェックエラー

var ErrPasswordLength = errors.New("password length is invalid")
var ErrPasswordForm = errors.New("password form is invalid")

// passwordのハッシュ化処理に失敗
var ErrPasswordFailedToHash = errors.New("hash to password failed")

// usernameの重複登録エラー
var ErrUsernameConflict = errors.New("username is conflict")

// 英数字を含むかどうかのバリデーションチェックのための関数
func containsAlphaAndNumber(s string) bool {
	hasAlpha := false
	hasDigit := false

	for _, r := range s {
		switch {
		case 'A' <= r && r <= 'Z':
			hasAlpha = true
		case 'a' <= r && r <= 'z':
			hasAlpha = true
		case '0' <= r && r <= '9':
			hasDigit = true
		}
		if hasAlpha && hasDigit {
			return true
		}
	}
	return false
}

func CreateUser(ctx context.Context, queries *database.Queries, input CreateUserInput) (CreateUserOutput, error) {
	//サインアップ処理の中身
	//バリデーション
	//usernameのバリデーションチェック
	if input.Username == "" {
		return CreateUserOutput{}, ErrUsernameRequired
	}

	if len(input.Username) < 5 || len(input.Username) > 15 {
		return CreateUserOutput{}, ErrUsernameLength
	}

	//passwordのバリデーションチェック
	if input.Password == "" {
		return CreateUserOutput{}, ErrPasswordRequired
	}

	if len(input.Password) < 8 || len(input.Password) > 64 {
		return CreateUserOutput{}, ErrPasswordLength
	}

	if !containsAlphaAndNumber(input.Password) {
		return CreateUserOutput{}, ErrPasswordForm
	}

	//user_idの払出。
	userID := uuid.New()

	//passwordのハッシュ化、ハッシュパスワードを文字列に変換
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Failed to Create HashPassword")
		return CreateUserOutput{}, ErrPasswordFailedToHash
	}
	passwordHashString := string(passwordHash)

	//repositoryのcreateuser関数を呼び出してDB処理をしてもらう
	repositoryInput := repository.CreateUserInput{
		UserID:       userID,
		Username:     input.Username,
		PasswordHash: passwordHashString,
	}

	repositoryOutput, err := repository.CreateUser(ctx, queries, repositoryInput)
	if errors.Is(err, repository.ErrUsernameConflict) {
		return CreateUserOutput{}, ErrUsernameConflict
	}

	if errors.Is(err, repository.ErrDB) {
		return CreateUserOutput{}, ErrDB
	}

	if err != nil {
		return CreateUserOutput{}, ErrDB
	}

	//すべてのチェックやDBへの格納やcreatedatの受け取りが通った後、CreateUserOutputに情報を詰める
	output := CreateUserOutput{
		UserID:    userID,
		Username:  input.Username,
		CreatedAt: repositoryOutput.CreatedAt,
	}

	return output, nil
}
