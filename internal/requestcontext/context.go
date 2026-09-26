package requestcontext

import (
	"context"

	"github.com/google/uuid"
)

//Context専用の出入り口を作る

// "user_id" という値を持つ、contextKey 型の定数 userIDKey を作る
type contextKey string

const userIDKey contextKey = "user_id"

// MiddlewareがContextにuserIDを格納する
// 既存のContextにuserIDを追加した、新しいContextを返す
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// HandlerがContextからuserIDを取り出す
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
