package service

import "errors"

//必須入力項目について汎用的なものなので共通のエラーとして定義
var ErrUsernameRequired = errors.New("username is required")
var ErrPasswordRequired = errors.New("password is required")
var ErrOriginalURLRequired = errors.New("original url is required")
var ErrShortCodeRequired = errors.New("shortcode is required")

// 汎用的なDBエラーを定義
var ErrDB = errors.New("database operation failed")
