package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	. "github.com/violin0622/grpc-apple/api/operation"
	"github.com/violin0622/grpc-apple/service"
)

func main() {
	serverCred, err := credentials.NewServerTLSFromFile(`./server.pem`, `./server.key`)
	if err != nil {
		log.Fatal(err)
	}
	clientCred, err := credentials.NewClientTLSFromFile(`./server.pem`, `localhost`)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(serverCred))
	RegisterAppleServiceServer(grpcServer, &service.AppleService{})
	reflection.Register(grpcServer)

	httpServer := runtime.NewServeMux()
	RegisterAppleServiceHandlerFromEndpoint(
		context.Background(),
		httpServer,
		`:8000`,
		[]grpc.DialOption{grpc.WithTransportCredentials(clientCred)},
	)

	http.ListenAndServeTLS(`:8000`, `./server.pem`, `./server.key`,
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
	)
}
