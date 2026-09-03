# POST /users

## Request Body:
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------| --------|   
| username | string | 5 |15 | 必須 | |
| password | string | 8 |64 | 必須,英数字をどちらも含む | |


## Response Body:
### 成功時
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------|--------|  
| user_id | string | - |- |必須|サーバ側で自動払出(uuid)| 
| username | string | 5 |15 | 必須 |リクエストで受けとった内容 |
| created_at |string | - | - | 必須 | サーバ側で自動払出(date-time)|

### 失敗時
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------| --------|   
| username | string | 5 |15 | 必須 |リクエストで受けとった内容 |
| error_message | string | 1 |256 | 必須 | "invalid username or password" |


## 認証:
 不要

## 成功時Status Code:
201

## 想定するエラー:
400 Bad Request：無効なデータの登録や必須項目がないこと
409 Conflict：usernameがすでに存在している
500 Internal Server Error：サーバ側での処理エラー
