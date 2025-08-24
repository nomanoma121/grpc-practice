package main

import (
	// (一部抜粋)
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	todopb "todo-grpc/gen/todo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type todoServiceServer struct {
	todopb.UnimplementedTodoServiceServer
}

type Todo struct {
	ID        string
	Title     string
	Completed bool
}

func main() {
	// 1. 8080番portのListenerを作成
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// 2. gRPCサーバーを作成
	s := grpc.NewServer()

	todopb.RegisterTodoServiceServer(s, &todoServiceServer{})

	reflection.Register(s)

	// 3. 作成したgRPCサーバーを、8080番ポートで稼働させる
	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()

	// 4.Ctrl+Cが入力されたらGraceful shutdownされるようにする
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}
