package main

import (
	"log"
	"net"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/verbovyar/OzonCart/api/CartServiceApiPb"
	"github.com/verbovyar/OzonCart/config"
	"github.com/verbovyar/OzonCart/internal/handlers"
	"github.com/verbovyar/OzonCart/internal/middleware"
	"github.com/verbovyar/OzonCart/internal/repositories/db/postgres"
	"github.com/verbovyar/OzonCart/internal/repositories/interfaces"
	"github.com/verbovyar/OzonCart/internal/service"
	"github.com/verbovyar/OzonCart/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// @title           Cart Service
// @version         1.0
// @description     HTTP сервис корзины. Стандартная библиотека, Postgres, валидация, ретраи к ProductService.
// @BasePath        /
// @schemes         http
func main() {
	//docs.SwaggerInfo.BasePath = "/"

	conf, err := config.LoadConfig("./config")
	if err != nil {
		println(err.Error())
	}
	log.Printf("%s", conf.ConnectingString)

	postgresStore := RunPostgres(conf.ConnectingString)
	cartService := RunService(conf.ProductURL, conf.ProductToken, postgresStore)
	RunGrpc(cartService, conf.Port, conf.NetworkType)
	//RunHttp(cartService, conf.Port)
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

func RunGrpc(cs *service.CartService, port string, networkType string) {
	listener, err := net.Listen(networkType, port)
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv := grpc.NewServer()
	CartServiceApiPb.RegisterCartServiceServer(grpcSrv, handlers.NewGrpsRouter(cs))

	// health + reflection
	healthpb.RegisterHealthServer(grpcSrv, health.NewServer())
	reflection.Register(grpcSrv)

	log.Println("CartService gRPC :50051")
	log.Fatal(grpcSrv.Serve(listener))
}
