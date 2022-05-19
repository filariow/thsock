package main

import (
	"log"
	"net"
	"os"

	"github.com/filariow/thsock/internal/thgrpc"
	"github.com/filariow/thsock/pkg/thprotos"
	"google.golang.org/grpc"
)

const SockAddr = "/tmp/th.socket"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Printf("Setting up socket file at '%s'\n", SockAddr)
	if err := os.RemoveAll(SockAddr); err != nil {
		return err
	}

	log.Printf("Listening on socket '%s'\n", SockAddr)
	ls, err := net.Listen("unix", SockAddr)
	if err != nil {
		return err
	}

	log.Println("Setting up gRPC server")
	s := grpc.NewServer()
	t := thgrpc.New()
	thprotos.RegisterTempHumSvcServer(s, t)

	log.Println("Startup process completed, waiting fpr gRPC requests...")
	return s.Serve(ls)
}
