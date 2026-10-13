package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hiroki-pixel/url-shortener/handler"
	"github.com/hiroki-pixel/url-shortener/internal/database"
	"github.com/hiroki-pixel/url-shortener/internal/requestcontext"
	"github.com/hiroki-pixel/url-shortener/repository"
	"github.com/hiroki-pixel/url-shortener/service"
	"github.com/joho/godotenv"
)

// コンテキスト内でキーの名前衝突を避けるためにstringをもとに自分専用の新しい型を作る
// authMiddleware及び認証が必要なhandler関数で利用するので、パッケージスコープにて宣言
type contextKey string

const userIDKey contextKey = "user_id"

// 認証が必要な機能で使用する、認証+目的の機能に処理を飛ばす関数
func authMiddleware(queries *database.Queries, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//認証処理内容。ここが正常にいけば、目的の機能に飛ばすし、でなければ401や500エラーを返す。
		//cookieの取得が実施できるかどうか
		sessionCookie, err := r.Cookie("session_id")
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}
		//cookie取得時にその他エラー発生時のエラーハンドリング
		if err != nil {
			http.Error(w, "Cookieの取得に失敗しました", http.StatusBadRequest)
			return
		}

		sessionID := sessionCookie.Value

		//sessionを取得するメソッドの使用
		getSession, err := queries.GetSessionByID(r.Context(), sessionID)

		// sqlcのGetSessionByIDから返されたDBエラーを判定する
		if errors.Is(err, sql.ErrNoRows) {
			//リクエストで受けとったsessionが存在しなかった
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}
		//その他のエラーが発生していないかの確認
		if err != nil {
			http.Error(w, "DBエラーが発生しました。時間をおいて再度試してください", http.StatusInternalServerError)
			return
		}

		//取得したセッションの有効期限を確認。
		if time.Now().After(getSession.ExpiresAt) {
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}

		//認証が完了したので、リクエストのコンテキストにuser_idを追加

		//コンテキストへの追加の実施
		ctx := requestcontext.WithUserID(r.Context(), getSession.UserID)

		//作成した新たなコンテキストをリクエストに紐づける
		r = r.WithContext(ctx)

		//新しいコンテキストが紐づけられたリクエストを後続Handlerに渡す。
		next.ServeHTTP(w, r)

		/* タイムゾーントラブル分析のため一時的に追加
		log.Printf(
			"now=%v, expiresAt=%v",
			time.Now(),
			getSession.ExpiresAt,
		)　*/

	})
}

func main() {

	//環境変数の取得確認
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error Loading .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not found")
	}
	log.Println("PORT:", port)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("baseURL is not found")
	}

	//DBへの接続確認
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB接続先の環境変数を確認してください。")
	}

	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("DB Open不可")
	}
	if err := db.Ping(); err != nil {
		log.Fatal("DB Connect不可")
	}
	fmt.Println("DB接続完了")

	//DBとの接続窓口たるdbを使って、sqlcで作成したメソッドを持つqueriesにする。
	queries := database.New(db)

	//Routerの作成
	mux := http.NewServeMux()

	//作成したRouter(つまり、mux)にhandlerを登録する
	//Post /usersのルーティング
	//
	userRepository := repository.NewUserRepository(queries)
	userService := service.NewUserService(userRepository)
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleCreateUser(w, r, userService)
	})
	//Post /sessionsのルーティング
	mux.HandleFunc("POST /sessions", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleCreateSession(w, r, queries)
	})

	//GET /{short_code}のルーティング
	mux.HandleFunc("GET /{short_code}", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleRedirect(w, r, queries)
	})

	//認証後に呼びたいHandler関数の用意をする。つまり、POST /urlsとGET /urls
	// POST / urlsに関してbaseURLが必須になるので引数に追加
	createURLHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.HandleCreateURL(w, r, queries, baseURL)
	})

	listURLHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.HandleListURL(w, r, queries, baseURL)
	})

	// authMiddleware(認証が必要なPOST /urls, GET /urls の利用のために必要)のルーティング
	mux.Handle("POST /urls", authMiddleware(queries, createURLHandler))
	mux.Handle("GET /urls", authMiddleware(queries, listURLHandler))

	//設定したアドレスとRouterを使用してサーバを起動
	addr := ":" + port
	log.Fatal(http.ListenAndServe(addr, mux))

}
