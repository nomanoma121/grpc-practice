package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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
		fmt.Println("5: Watch todos (realtime)")
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
			fmt.Print("Enter todo title: ")
			var title string
			_, err := fmt.Scanf("%s", &title)
			if err != nil {
				fmt.Println("Invalid input. Please enter a title.")
				continue
			}
			handler.Create(context.Background(), title)
		case 3:
			fmt.Print("Enter todo ID to update: ")
			var id string
			_, err := fmt.Scanf("%s", &id)
			if err != nil {
				fmt.Println("Invalid input. Please enter a valid ID.")
				continue
			}
			// どの項目を変更したいですか？
			// 1: title, 2: completed, 3: both, 4: back
			fmt.Println("What do you want to update?")
			fmt.Println("1: title")
			fmt.Println("2: completed")
			fmt.Print("Enter your choice: ")
			var updateChoice int
			_, err = fmt.Scanf("%d", &updateChoice)
			if err != nil {
				fmt.Println("Invalid input. Please enter a number.")
				continue
			}
			switch updateChoice {
			case 1:
				fmt.Print("Enter new title: ")
				var title string
				_, err = fmt.Scanf("%s", &title)
				if err != nil {
					fmt.Println("Invalid input. Please enter a title.")
					continue
				}
				handler.Update(context.Background(), id, &title, nil)
			case 2:
				fmt.Print("Is it completed? (y/n): ")
				var completed string
				_, err = fmt.Scanf("%s", &completed)
				if err != nil {
					fmt.Println("Invalid input. Please enter (y/n).")
					continue
				}
				var isCompleted bool
				if completed == "y" {
					isCompleted = true
				} else if completed == "n" {
					isCompleted = false
				} else {
					fmt.Println("Invalid input. Please enter (y/n).")
					continue
				}
				handler.Update(context.Background(), id, nil, &isCompleted)
			default:
				fmt.Println("Invalid choice. Please try again.")
			}
		case 4:
			fmt.Print("Enter todo ID to delete: ")
			var id string
			_, err := fmt.Scanf("%s", &id)
			if err != nil {
				fmt.Println("Invalid input. Please enter a valid ID.")
				continue
			}
			handler.Delete(context.Background(), id)
		case 5:
			fmt.Println("Watching todos ...")
			handler.Watch(context.Background())
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
	fmt.Println("---------------------------------------")
	for _, todo := range r.Todos {
		fmt.Printf("- ID: %s, Title: %s, Completed: %v\n", todo.Id, todo.Title, todo.Completed)
	}
	fmt.Println("---------------------------------------")
}

func (h *TodoHandler) Create(ctx context.Context, title string) {
	r, err := h.client.CreateTodo(ctx, &todopb.CreateTodoRequest{Title: title})
	if err != nil {
		log.Printf("could not create todo: %v", err)
		return
	}
	fmt.Printf("Created todo: ID: %s, Title: %s, Completed: %v\n", r.Id, r.Title, r.Completed)
}

func (h *TodoHandler) Update(ctx context.Context, id string, title *string, completed *bool) {
	r, err := h.client.UpdateTodo(ctx, &todopb.UpdateTodoRequest{
		Id:        id,
		Title:     title,
		Completed: completed,
	})
	if err != nil {
		log.Printf("could not update todo: %v", err)
		return
	}
	fmt.Printf("Updated todo: ID: %s, Title: %s, Completed: %v\n", r.Id, r.Title, r.Completed)
}

func (h *TodoHandler) Delete(ctx context.Context, id string) {
	_, err := h.client.DeleteTodo(ctx, &todopb.DeleteTodoRequest{Id: id})
	if err != nil {
		log.Printf("could not delete todo: %v", err)
		return
	}
	fmt.Printf("Deleted todo: ID: %s\n", id)
}

func (h *TodoHandler) Watch(ctx context.Context) {
	stream, err := h.client.WatchTodos(ctx)
	if err != nil {
		log.Printf("could not watch todos: %v", err)
		return
	}
	defer stream.CloseSend()

	err = stream.Send(&todopb.WatchTodosRequest{
		Action: &todopb.WatchTodosRequest_Start{
			Start: &todopb.StartWatchRequest{},
		},
	})

	if err != nil {
		log.Printf("could not send watch request: %v", err)
		return
	}

	go func() {
		fmt.Println("Press Enter to stop watching...")
		fmt.Scanln()
		stream.Send(&todopb.WatchTodosRequest{
			Action: &todopb.WatchTodosRequest_Stop{
				Stop: &todopb.StopWatchRequest{},
			},
		})
	}()

	for {
		todo, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error receiving todo update: %v", err)
			break
		}
		fmt.Printf("Received todo update: ID: %s, Title: %s, Completed: %v\n", todo.Id, todo.Title, todo.Completed)
	}
}
