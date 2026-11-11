package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hiroki-pixel/url-shortener/repository"
)

func TestAuthenticateSession(t *testing.T) {
	// Arrange
	wantUserID := uuid.New()
	fakeRepo := &FakeSessionRepository{
		UserID:    wantUserID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	svc := NewSessionService(fakeRepo)

	ctx := context.Background()

	// inputを定義する
	input := AuthenticateSessionInput{
		SessionID: "test-session-id",
	}

	// Act
	// AuthenticateSessionを実行する
	got, err := svc.AuthenticateSession(ctx, input)

	// Assert
	// 結果を検証する
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if got.UserID != wantUserID {
		t.Errorf("got %v, want %v", got.UserID, wantUserID)
	}

}

func TestAuthenticateSessionExpired(t *testing.T) {
	// Arrange
	wantUserID := uuid.New()

	fakeRepo := &FakeSessionRepository{
		UserID:    wantUserID,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	svc := NewSessionService(fakeRepo)

	ctx := context.Background()

	input := AuthenticateSessionInput{
		SessionID: "test-session-id",
	}

	// Act
	_, err := svc.AuthenticateSession(ctx, input)

	// Assert
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}

	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("got %v,want %v", err, ErrSessionExpired)
	}
}

func TestAuthenticateSessionNotFound(t *testing.T) {
	// Arrange
	fakerepo := &FakeSessionRepository{
		ReturnErr: repository.ErrSessionNotFound,
	}

	svc := NewSessionService(fakerepo)
	ctx := context.Background()

	input := AuthenticateSessionInput{
		SessionID: "test-session-id",
	}

	// Act
	_, err := svc.AuthenticateSession(ctx, input)

	// Assert
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("got %v,want %v", err, ErrSessionNotFound)
	}
}

func TestAuthenticateSessionDBError(t *testing.T) {
	// Arrange
	fakeRepo := &FakeSessionRepository{
		ReturnErr: errors.New("database connection failed"),
	}

	svc := NewSessionService(fakeRepo)

	// ContextとInputを準備
	ctx := context.Background()
	input := AuthenticateSessionInput{
		SessionID: "DBtest-session-id",
	}
	// Act
	// AuthenticateSession()を実行
	_, err := svc.AuthenticateSession(ctx, input)

	// Assert
	// 期待するエラーか検証
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}

	if !errors.Is(err, ErrDB) {
		t.Errorf("got %v, want %v", err, ErrDB)
	}
}
