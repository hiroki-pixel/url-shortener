package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestAuthenticateSession(t *testing.T) {
	// Arrange
	wantUserID := uuid.New()
	fakeRepo := &FakeSessionRepository{
		UserID: wantUserID,
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
