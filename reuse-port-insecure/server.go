package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	. "github.com/violin0622/grpc-apple/api/operation"
	"github.com/violin0622/grpc-apple/service"
)

func main() {
	grpcServer := grpc.NewServer()
	RegisterAppleServiceServer(grpcServer, &service.AppleService{})
	reflection.Register(grpcServer)

	httpServer := runtime.NewServeMux()
	RegisterAppleServiceHandlerFromEndpoint(
		context.Background(),
		httpServer,
		`:8000`,
		[]grpc.DialOption{grpc.WithInsecure()},
	)

	http.ListenAndServe(
		`:8000`,
		h2c.NewHandler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				if r.ProtoMajor == 2 &&
					strings.Contains(r.Header.Get(`Content-Type`), `application/grpc`) {
					log.Println(`grpc`)
					grpcServer.ServeHTTP(w, r)
				} else {
					log.Println(`http`)
					httpServer.ServeHTTP(w, r)
				}
			}),
			&http2.Server{}),
	)
}
