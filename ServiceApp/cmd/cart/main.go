package main

import (
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/verbovyar/OzonCart/config"
	"github.com/verbovyar/OzonCart/internal/handlers"
	"github.com/verbovyar/OzonCart/internal/middleware"
	"github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
	"github.com/verbovyar/OzonCart/internal/repositories/interfaces"
	"github.com/verbovyar/OzonCart/internal/service"
	"github.com/verbovyar/OzonCart/pkg"
)

func main() {
	conf, err := config.LoadConfig("./config")
	if err != nil {
		println(err.Error())
	}

	postgresStore := RunPostgres(conf.ConnectingString)
	cartService := RunService(conf.ProductURL, conf.ProductToken, postgresStore)
	RunHttp(cartService, conf.HttpPort, conf.SwaggerPort)
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

func RunHttp(cs *service.CartService, http_port, swager_port string) {
	mux := http.NewServeMux()
	mux.Handle("/user/", handlers.New(cs))           // handlers
	mux.Handle("/swagger/", httpSwagger.WrapHandler) // swagger

	handlerWithMW := middleware.Logging(mux)

	log.Printf("HTTP server is listening on port%s", http_port)
	log.Printf("Swagger UI: http://localhost%s/swagger/index.html", swager_port)

	err := http.ListenAndServe(http_port, handlerWithMW)
	if err != nil {
		log.Fatal(err)
	}
}
