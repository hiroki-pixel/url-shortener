# GET /urls

## Resquest Header
### Request cookie　
 session_id：必須（POST /sessionsで発行されたセッションID）


## Request Body:
空

## Response Body:
### 成功時
URL情報の配列となる。
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------| --------|   
| short_code | string | 6 | | 6 | 必須 | 登録済みのshort_code|
| short_url | string | - | | - | 必須 | 登録済みのshort_url|
| original_url | string | 1 |2048 | 必須 | 登録済みのoriginal_url|
| memo | string | 0 |256 | 任意|登録済みのmemo|

登録がゼロ件の場合は、[]として返却



### 失敗時
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------|--------|  
| error_message | string | 1 | 256 |必須| エラー内容 | 

## 認証:
必要

## 成功時Status Code:
200
[] (対象ユーザが登録しているurlが0件であったとき)

## 想定するエラー:
401 Unauthorized：認証失敗
500 Internal Server Error：サーバ側での処理エラー
