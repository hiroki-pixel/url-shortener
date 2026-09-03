# POST /sessions

## Request Body:
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------| --------|   
| username | string | 5 |15 | 必須 | |
| password | string | 8 |64 | 必須,英数字をどちらも含む | |

## Response Header
### cookie
cookie名：session_id
有効期限：10800秒(3時間)
HTTP Only：true (JavaScriptからのアクセスを禁止する)
Secure：true(HTTPS通信の時のみcookieを送信する)
※ローカルでは「http://localhost:8080」使うので、falseにする。
環境変数によって切り替える設定にする
Same Site：Lax(別のサイトから飛んできたリクエストにも、自分のCookieを付けるのか？一部の安全な遷移では許可する設定)

## Response Body:
### 成功時
Response Bodyは空

### 失敗時
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------|--------|  
| error_message | string | 1 | 256 |必須| "invalid username or password" | 

## 認証:
不要

## 成功時Status Code:
201

## 想定するエラー:
400 Bad Request：無効なデータの登録や必須項目がないこと
401 Unauthorized：認証失敗
500 Internal Server Error：サーバ側での処理エラー
