package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	todopb "todo-grpc/gen/todo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	addr = flag.String("addr", "localhost:8080", "the address to connect to")
)

func main() {
	flag.Parse()
	
	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := todopb.NewTodoServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
	for {
		fmt.Println("\nWhat do you want to do?")
		fmt.Println("1: Get todos")
		fmt.Println("2: Create todo")
		fmt.Println("3: Update todo")
		fmt.Println("4: Delete todo")
		fmt.Println("0: Exit")
		fmt.Print("Enter your choice: ")
		
		var choice int
		_, err := fmt.Scanf("%d", &choice)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}
		
		switch choice {
		case 1:
			getTodos(ctx, c)
		case 2:
			fmt.Println("Create todo - Not implemented yet")
		case 3:
			fmt.Println("Update todo - Not implemented yet")
		case 4:
			fmt.Println("Delete todo - Not implemented yet")
		case 0:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func getTodos(ctx context.Context, c todopb.TodoServiceClient) {
	r, err := c.GetTodos(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("could not get todos: %v", err)
	}
	
	fmt.Printf("Retrieved %d todos:\n", len(r.Todos))
	for _, todo := range r.Todos {
		fmt.Printf("- ID: %s, Title: %s, Completed: %v\n", todo.Id, todo.Title, todo.Completed)
	}
}
