# API一覧

## POST /users
ユーザ登録を実施する

## POST /sessions
ユーザ認証を行い、セッションIDを発行する

## POST /urls
認証済みユーザが短縮したいURLを送付し、短縮版URLを作成して返却する

## GET /urls
認証済みユーザが作成したURLの一覧を表示する

## GET /url/{short_code}
短縮版urlから元のURLへリダイレクトを行う