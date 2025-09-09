package main

import (
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/verbovyar/OzonCart/internal/handlers"
	"github.com/verbovyar/OzonCart/internal/middleware"
	"github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
	"github.com/verbovyar/OzonCart/internal/repositories/interfaces"
	"github.com/verbovyar/OzonCart/internal/service"
	"github.com/verbovyar/OzonCart/pkg"
)

func main() {
	postgresStore := RunPostgres("")
	cartService := RunService("", "", postgresStore)
	RunHttp(cartService)
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

func RunHttp(cs *service.CartService) {
	mux := http.NewServeMux()
	mux.Handle("/user/", handlers.New(cs))           // handlers
	mux.Handle("/swagger/", httpSwagger.WrapHandler) // swagger

	handlerWithMW := middleware.Logging(mux)

	log.Printf("HTTP server is listening on %s port: ", "")
	log.Printf("Swagger UI: http://localhost%s/swagger/index.html", "")

	err := http.ListenAndServe("", handlerWithMW)
	if err != nil {
		log.Fatal(err)
	}
}
