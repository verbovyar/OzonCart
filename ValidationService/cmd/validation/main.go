// ValidationService/cmd/validation/main.go
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	valpb "validation/api/ValidationServiceApiPb"
	"validation/handlers"
	cartpb "validation/infrastructure/cartServiceClient/api/CartServiceApiPb"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cartConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	cartClient := cartpb.NewCartServiceClient(cartConn)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}
	grpcSrv := grpc.NewServer()
	valSrv := handlers.NewValidationRouter(cartClient)
	valpb.RegisterValidationServiceServer(grpcSrv, valSrv)

	go func() {
		log.Println("ValidationService gRPC listening :50052")
		log.Fatal(grpcSrv.Serve(lis))
	}()

	gwMux := gw.NewServeMux()
	if err := valpb.RegisterValidationServiceHandlerServer(context.Background(), gwMux, valSrv); err != nil {
		log.Fatal(err)
	}

	swaggerJSON, err := os.ReadFile("../../ValidationService/api/ValidationService.swagger.json")
	if err != nil {
		log.Fatalf("cannot read swagger file: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusFound)
	})

	mux.Handle("/api/", http.StripPrefix("/api", gwMux))

	mux.HandleFunc("/swagger/ValidationService.swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(swaggerJSON)
	})

	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!doctype html>
<html>
<head>
  <meta charset="utf-8"/>
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/swagger/ValidationService.swagger.json',
      dom_id: '#swagger-ui'
    })
  </script>
</body>
</html>`))
	})

	log.Println("ValidationService HTTP (gateway + swagger) :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
