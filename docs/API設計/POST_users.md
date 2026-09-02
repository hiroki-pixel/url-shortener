# POST /users

## Request Body:
| 物理名 | 型 |最小桁数|最大桁数|制約|
 |--------|--------|--------|--------|  
| username | string | 5 |15 | 必須 |
| password | string | 7 |15 | 必須,英数字をどちらも含む |


## Response Body:
| 物理名 | 型 |最小桁数|最大桁数|制約|
 |--------|--------|--------|--------| 
| user_id | uuid | - |- |自動払出、必須 | 
| username | string | 5 |15 | 必須 |
| password | string | 7 |15 | 必須,英数字をどちらも含む |


## 認証:
 不要

## 成功時Status Code:
201

## 想定するエラー:
400 Bad Request：無効なデータの登録や必須項目がないこと
409 Conflict：usernameがすでに存在している
500 Internal Server Error：サーバ側での処理エラー