package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	. "github.com/violin0622/grpc-apple/api/operation"
	"github.com/violin0622/grpc-apple/service"
)


func main() {
	grpcServer := grpc.NewServer()
	RegisterAppleServiceServer(grpcServer, &service.AppleService{})
	reflection.Register(grpcServer)

	if l, err := net.Listen(`tcp`, `:9000`); err != nil {
		log.Fatal(`cannot listen to port 9000: `, err)
	} else if err = grpcServer.Serve(l); err != nil {
		log.Fatal(`cannot start service:`, err)
	}
}
