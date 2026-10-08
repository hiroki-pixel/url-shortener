package repository

import "errors"

// 汎用的なDBエラーを定義
var ErrDB = errors.New("database operation failed")
