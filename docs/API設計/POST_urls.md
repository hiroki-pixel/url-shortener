# POST /urls

## Resquest Header
### Request cookie　
 session_id：必須（POST /sessionsで発行されたセッションID）


## Request Body:
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------| --------|   
| original_url | string | 1 |2048 | 必須 | 短縮対象にしたいurl。urlとしての形式チェックを実施する。User_idと組み合わせ一意である必要あり。|
| memo | string | 0 |256 | 任意|ユーザが用途とかメモを残したければ残せるようにする |



## Response Body:
### 成功時
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------| --------|   
| short_code | string | 6 | | 6 | 必須 | サーバ側でランダム生成した一意の英数字|
| short_url | string | - | | - | 必須 | short_codeを含む短縮版url|



### 失敗時
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------|--------|  
| error_message | string | 1 | 256 |必須| エラー内容 | 

## 認証:
必要

## 成功時Status Code:
201

## 想定するエラー:
400 Bad Request：無効なデータの登録や必須項目がないこと、URLとしての形式不正
409 Conflict： user_id + original_urlの重複時にエラー返却
401 Unauthorized：認証失敗
500 Internal Server Error：サーバ側での処理エラー


