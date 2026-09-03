# GET /{short_code}

## Resquest Header


## Request Body:
なし

## Path Parameter:
 short_code：
- string 
- 6文字 
- 必須 
- 英数字

## Response Header:
Location：original_url

## Response Body:
なし

### 成功時
返したいJSON情報はない
ブラウザがoriginal_urlへリダイレクトされる


### 失敗時
| 物理名 | 型 |最小桁数|最大桁数|制約|備考|
 |--------|--------|--------|--------|--------|--------|  
| error_message | string | 1 | 256 |必須| エラー内容 | 

## 認証:
不要

## 成功時Status Code:
302 Found


## 想定するエラー:
404 Not Found：short_codeが検索にひっかからなかった
500 Internal Server Error：サーバ側での処理エラー
