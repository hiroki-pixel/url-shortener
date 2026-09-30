package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/lib/pq"
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

	//DBへのユーザ情報格納、CreatedAtの取得
	insertParams := database.InsertUserParams{
		UserID:       userID,
		Username:     input.Username,
		PasswordHash: passwordHashString,
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

	//CreateUserOutputに情報を詰める
	output := CreateUserOutput{
		UserID:    userID,
		Username:  input.Username,
		CreatedAt: createdAt,
	}

	return output, nil
}
