// ValidationService/cmd/validation/main.go
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	valpb "validation/api/ValidationServiceApiPb"
	"validation/handlers"
	cartpb "validation/infrastructure/cartServiceClient/api/CartServiceApiPb"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func mustReadSwagger() []byte {
	wd, _ := os.Getwd()

	candidates := []string{
		filepath.Join(wd, "ValidationService/api/ValidationService.swagger.json"),
		filepath.Join(wd, "api/ValidationService.swagger.json"),
		filepath.Join(wd, "../../ValidationService/api/ValidationService.swagger.json"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			log.Printf("Swagger loaded from %s", path)
			return data
		}
	}

	log.Fatalf("Swagger file not found. Tried:\n%s", strings.Join(candidates, "\n"))
	return nil
}

func main() {
	// --- gRPC-клиент CartService (не падаем в «подвис», добавим WithBlock + таймаут) ---
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cartConn, err := grpc.DialContext(
		ctx,
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("dial cart failed: %v", err)
	}
	defer cartConn.Close()

	cartClient := cartpb.NewCartServiceClient(cartConn)

	// --- gRPC-сервер ValidationService на :50052 ---
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

	// --- HTTP Gateway (ВАЖНО: вешаем на КОРЕНЬ, без /api префиксов) ---
	gwMux := gw.NewServeMux()
	if err := valpb.RegisterValidationServiceHandlerServer(context.Background(), gwMux, valSrv); err != nil {
		log.Fatal(err)
	}

	// --- Swagger JSON + простой UI ---
	swaggerJSON := mustReadSwagger()

	mux := http.NewServeMux()

	// Swagger JSON
	mux.HandleFunc("/swagger/ValidationService.swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(swaggerJSON)
	})

	// Простой Swagger UI (CDN), чтобы «Try it out» бил в gateway
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
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

	// Корневой обработчик:
	// - только точный GET "/" редиректим на UI
	// - все остальные пути отдаём gateway (чтобы /validate/... шёл в хендлеры)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == http.MethodGet {
			http.Redirect(w, r, "/swagger", http.StatusFound)
			return
		}
		gwMux.ServeHTTP(w, r)
	})

	log.Println("ValidationService HTTP (gateway + swagger) :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
