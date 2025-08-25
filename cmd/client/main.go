package main

import (
	"context"
	"flag"
	"fmt"
	"log"

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

	handler := NewTodoHandler(c)

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
			handler.Get(context.Background())
		case 2:
			// Titleを入力させる
			fmt.Print("Enter todo title: ")
			var title string
			_, err := fmt.Scanf("%s", &title)
			if err != nil {
				fmt.Println("Invalid input. Please enter a title.")
				continue
			}
			handler.Create(context.Background(), title)
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

type TodoHandler struct {
	client todopb.TodoServiceClient
}

func NewTodoHandler(client todopb.TodoServiceClient) *TodoHandler {
	return &TodoHandler{
		client: client,
	}
}

func (h *TodoHandler) Get(ctx context.Context) {
	r, err := h.client.GetTodos(ctx, &emptypb.Empty{})
	if err != nil {
		log.Printf("could not get todos: %v", err)
		return
	}
	
	fmt.Printf("Retrieved %d todos:\n", len(r.Todos))
	for _, todo := range r.Todos {
		fmt.Printf("- ID: %s, Title: %s, Completed: %v\n", todo.Id, todo.Title, todo.Completed)
	}
}

func (h *TodoHandler) Create(ctx context.Context, title string) {
	r, err := h.client.CreateTodo(ctx, &todopb.CreateTodoRequest{Title: title})
	if err != nil {
		log.Printf("could not create todo: %v", err)
		return
	}
	fmt.Printf("Created todo: ID: %s, Title: %s, Completed: %v\n", r.Id, r.Title, r.Completed)
}
