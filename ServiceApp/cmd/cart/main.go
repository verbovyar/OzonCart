package main

import (
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/verbovyar/OzonCart/config"
	"github.com/verbovyar/OzonCart/internal/docs"
	"github.com/verbovyar/OzonCart/internal/handlers"
	"github.com/verbovyar/OzonCart/internal/middleware"
	"github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
	"github.com/verbovyar/OzonCart/internal/repositories/interfaces"
	"github.com/verbovyar/OzonCart/internal/service"
	"github.com/verbovyar/OzonCart/pkg"
)

// @title           Cart Service
// @version         1.0
// @description     HTTP сервис корзины. Стандартная библиотека, Postgres, валидация, ретраи к ProductService.
// @BasePath        /
// @schemes         http
func main() {
	docs.SwaggerInfo.BasePath = "/"

	conf, err := config.LoadConfig("./config")
	if err != nil {
		println(err.Error())
	}

	postgresStore := RunPostgres(conf.ConnectingString)
	cartService := RunService(conf.ProductURL, conf.ProductToken, postgresStore)
	RunHttp(cartService, conf.Port)
}

func RunPostgres(connectionString string) *postgres.Store {
	p := pkg.New(connectionString)
	store := postgres.New(p.Pool)

	return store
}

func RunService(productURL, productToken string, store interfaces.RepositoryIface) *service.CartService {
	pc := service.NewClient(productURL, productToken, 3, 300*time.Millisecond)
	cs := service.New(store, pc)

	return cs
}

func RunHttp(cs *service.CartService, port string) {
	mux := http.NewServeMux()
	mux.Handle("/user/", handlers.New(cs)) // handlers

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	// ВАЖНО: URL("/swagger/doc.json"), и именно префикс /swagger/
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	handlerWithMW := middleware.Logging(mux)

	log.Printf("HTTP server is listening on port%s", port)
	log.Printf("Swagger UI: http://localhost%s/swagger/index.html", port)

	err := http.ListenAndServe(port, handlerWithMW)
	if err != nil {
		log.Fatal(err)
	}
}
