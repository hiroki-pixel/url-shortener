# M11 AWSアーキテクチャ設計メモ（構成図作成用）

更新日：2026-10-27

## 1. 目的と前提

- 対象：Goで実装したURL短縮サービス（`POST /users`、`POST /sessions`、`POST /urls`、`GET /urls`、`GET /{short_code}`）。
- 方針：**個人開発・小規模利用を想定し、コストと運用管理の手間をできるだけ抑える**。
- API処理は同期・短時間の軽量な処理であるため、コンピューティングにはAWS Lambdaを採用する。
- 既存のPostgreSQL / SQL / sqlcの実装を活用する。今回DynamoDBへの移行はスコープ外。
- 認証方式は既存どおりCookie + DB保存セッション方式。

## 2. サービス選定（現時点の方針）

| 領域 | 採用サービス・方式 | 理由・備考 |
|---|---|---|
| 静的フロントエンド | Amazon S3 + Amazon CloudFront | 静的ファイルを配信する |
| API受付 | Amazon API Gateway | インターネットからのHTTPS APIリクエストを受け付ける |
| API処理 | **AWS Lambda（Go）、1関数に全5 APIを集約** | 軽量・短時間の処理。デプロイ・IAM・設定・監視対象を最小限にする |
| データベース | Amazon RDS for PostgreSQL | 既存のSQLベースのアプリを活かす。Aurora Serverlessはコスト面から見送り |
| DB接続 | Lambda → RDSへの直接接続 | 小規模想定のため当初はRDS Proxyなし。接続数は監視対象 |
| シークレット | AWS Secrets Manager | DB認証情報などを管理する |
| ログ | Amazon CloudWatch Logs | Lambda標準ログを保存する |
| CI/CD | GitHub Actions | 既存GitHubリポジトリを利用。AWSへの認証はOIDC連携を想定 |
| 今回使わないもの | DynamoDB、RDS Proxy、Aurora Serverless、CodeCommit/CodeDeploy/CodePipeline | 学習スコープ・コスト・構成の単純さから初期構成では不採用 |

## 3. リクエストの流れ

### 3.1 静的画面とAPI

```text
ユーザーのブラウザ
  ├─ 画面取得 → CloudFront → S3（HTML / JS / CSS）
  └─ API呼び出し（画面上の操作など）→ API Gateway → Lambda
```

- **S3がAPI Gatewayへのリクエストを中継するわけではない**。ブラウザが画面のJavaScriptからAPIを呼び出す。
- フロントエンドとAPIを別ドメインとするか、CloudFrontで同じドメインに集約するかは未決定。Cookie・CORS・SameSiteの設計に関係する。

### 3.2 API GatewayとLambda

```text
API Gateway ── Lambda統合（Invoke）──→ Lambda
```

- API Gatewayのルートと対象Lambda関数を関連付ける。
- LambdaのリソースベースポリシーでAPI Gatewayからの`lambda:InvokeFunction`を許可する。
- **この呼び出しのためのVPC Linkは不要**。API GatewayからLambdaのプライベートIPへTCP接続する方式ではない。
- 現在のGoアプリは`net/http`ベースなので、Lambda実行形式に合わせたアダプター等の対応が必要。

### 3.3 Lambdaの関数分割方針（決定済み）

**初期構成では、5つのAPIを1つのLambda関数にまとめる。**

対象API：`POST /users`、`POST /sessions`、`POST /urls`、`GET /urls`、`GET /{short_code}`。

**採用理由**

- 個人開発・小規模利用であり、現時点ではAPIごとの独立デプロイや個別スケーリングの要件がない。
- Lambda関数・IAMロール・環境変数・監視・デプロイ設定の管理対象を減らし、運用負荷を抑える。
- 現在のGoアプリの`net/http`ルーティングと共通コードを活用しやすい。
- すべて同じRDS PostgreSQLを利用する。

**混同しないこと**：Lambdaを1関数に集約しても、Goコード内部の**Handler → Service → Repository**という責務分離は維持する。Lambdaの数（デプロイ単位）と、Goコードの層（設計単位）は別。

```text
API Gateway
    │ 5つのルートを1つの関数へ統合
    ▼
Lambda（1関数）
    ├─ Handler / Middleware
    ├─ Service
    └─ Repository
            │
            ▼
      RDS PostgreSQL
```

**将来の再検討条件**：例えば`GET /{short_code}`だけにアクセスが集中し、個別の同時実行制御、性能設定、デプロイ、IAM権限の分離が必要になった場合、リダイレクト機能だけを別Lambdaに切り出すことを検討する。なお、Lambdaを分割しても共有RDSの接続数・負荷は別途管理が必要。

## 4. VPC・ネットワーク設計

### 4.1 配置

- RDS PostgreSQL：自分のVPC内の**プライベートサブネット**に配置し、パブリックアクセスを無効にする。
- Lambda：**自分のVPCに接続**し、プライベートサブネットを指定する。
  - 厳密にはLambdaの実行環境そのものがサブネット内に置かれるというより、VPC接続用のネットワークインターフェース経由でVPCリソースへアクセスする。
- Lambda → RDSはVPC内のプライベート通信。NAT Gatewayやインターネットを経由しない。
- API GatewayはVPCの外からLambda統合で関数を呼び出す。

### 4.2 Security Group

| 対象 | 向き | 許可内容 |
|---|---|---|
| `rds-sg` | インバウンド | 送信元`lambda-sg`、TCP **5432** |
| `lambda-sg` | アウトバウンド | RDSへのTCP **5432**、Secrets Manager用VPC EndpointへのTCP **443** |
| `secretsmanager-vpce-sg`（仮称） | インバウンド | 送信元`lambda-sg`、TCP **443** |

- Security Groupはステートフルであり、許可した通信の戻りトラフィックについて逆向きの個別許可は通常不要。
- 送信元として指定するのはLambda関数名ではなく、**LambdaのVPC接続に設定したSecurity Group**。
- 最小限の通信許可を目標とする。実際のアウトバウンド設定は構築時に確認する。

### 4.3 Secrets Managerへのアクセス

```text
Lambda（VPC接続あり）
  → Interface VPC Endpoint（Secrets Manager）
  → AWS Secrets Manager
```

- NAT Gatewayは初期構成では採用せず、**Secrets Manager用Interface VPC Endpoint**を利用する案。
- EndpointのPrivate DNSを有効化し、通常のSecrets Manager API名でプライベート接続できるようにする。
- Lambdaの実行ロールに必要最小限の`secretsmanager:GetSecretValue`権限を付与する。必要に応じてKMS権限も確認する。
- Interface Endpointには**時間料金・データ処理料金**がある。構成確定前に月額コストを試算する。

### 4.4 CloudWatch Logsへのアクセス

- Goの`log.Println()`など、**Lambda標準のログ出力**を使用する。
- 標準ログの配送はLambdaサービス側が担うため、通常のLambdaであればCloudWatch Logs用VPC EndpointやNAT Gatewayを追加する必要はない。
- Lambda実行ロールには必要なCloudWatch Logs権限を付与する。
- **アプリ自身がAWS SDKでCloudWatch Logs APIを直接呼ぶ**場合は別であり、プライベート接続経路を検討する。

## 5. 構成図の下書き用・概念図

```text
 [利用者のブラウザ]
       │
       ├── 静的ファイル取得 ──→ [CloudFront] ──→ [S3]
       │
       └── HTTPS API呼出 ────→ [API Gateway]
                                        │ Lambda統合（Invoke）
                                        ▼
            ┌──────────── 自分のVPC ─────────────┐
            │ Private Subnet（VPC接続先）        │
            │                                   │
            │  [Lambda：5 APIを1関数に集約]       │
            │      │ TCP 5432         │           │
            │      └────────────────→ [RDS]     │
            │      │                   PostgreSQL│
            │      │ HTTPS 443                   │
            │      ▼                             │
            │ [Secrets Manager Interface VPCE] │
            └──────┬────────────────────────────┘
                   │
                   ▼
              [Secrets Manager]

 [Lambda標準ログ] ── Lambdaサービスのログ配送 ──→ [CloudWatch Logs]
 [GitHub] ──→ [GitHub Actions] ── OIDC / AWSデプロイ ──→ [Lambda等]
```

> 注：概念図であり、AZごとのサブネット数、RDSのSubnet Group、各サービスの厳密なAWS管理境界までは描いていない。

## 6. 構成図作成時に残っている検討事項

- [x] Lambdaは**1関数に全5 APIを集約**する（運用負荷を優先し、将来必要に応じて分割）。
- [ ] フロントとAPIを**同一ドメインに統合するか、別ドメインにするか**（Cookie、CORS、SameSite、Secure）。
- [ ] VPC内のAZ数・サブネット数、RDSのDB Subnet Group、Single-AZ構成の詳細。
- [ ] RDSインスタンスサイズ・ストレージ、Secrets ManagerとVPC Endpointを含む**月額料金の概算**。
- [ ] Lambda同時実行数、`database/sql`の接続プール設定、RDSの接続数監視。
- [ ] Secrets Managerの取得タイミングとシークレット更新方法。
- [ ] GitHub Actionsでのビルド、テスト、Lambdaデプロイ、および必要なAWS IAM権限。
- [ ] `net/http`のGoアプリをLambdaへ適合させる実装方法。
- [ ] CloudWatchの監視項目（Lambdaエラー・実行時間、API Gateway 5xxなど）。

## 7. 次の作業

1. この設計メモをもとにdraw.ioで**初版のAWS構成図**を作成する。
2. 採用理由・迷っている点を構成図にメモする。
3. 図のレビューで、**ネットワークの成立性、セキュリティ、コスト、運用の手間**を確認して改善する。

