package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	todopb "todo-grpc/gen/todo"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TodoStore struct {
	mu        sync.RWMutex
	todos     map[string]*todopb.Todo
	todoIndex int
}

func NewTodoStore() *TodoStore {
	return &TodoStore{
		todos:     make(map[string]*todopb.Todo),
		todoIndex: 0,
	}
}

func (ts *TodoStore) Get(id string) *todopb.Todo {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.todos[id]
}

func (ts *TodoStore) Add(todo *todopb.Todo) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.todos[todo.Id] = todo
}

func (ts *TodoStore) GetAll() []*todopb.Todo {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	todos := make([]*todopb.Todo, 0, len(ts.todos))
	for _, todo := range ts.todos {
		todos = append(todos, todo)
	}
	return todos
}

func (ts *TodoStore) Update(todo *todopb.Todo) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.todos[todo.Id] = todo
}

func (ts *TodoStore) Delete(id string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.todos, id)
}

type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]chan *todopb.Todo
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]chan *todopb.Todo),
	}
}

func (b *Broker) Subscribe() (string, chan *todopb.Todo) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := uuid.New().String()
	ch := make(chan *todopb.Todo)
	b.subscribers[id] = ch
	return id, ch
}

func (b *Broker) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscribers[id]; ok {
		delete(b.subscribers, id)
		close(ch)
	}
}

func (b *Broker) Publish(todo *todopb.Todo) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers {
		ch <- todo
	}
}

type todoServiceServer struct {
	todopb.UnimplementedTodoServiceServer
	store  *TodoStore
	broker *Broker
}

func (s *todoServiceServer) CreateTodo(ctx context.Context, req *todopb.CreateTodoRequest) (*todopb.Todo, error) {
	todo := &todopb.Todo{
		Id:        fmt.Sprintf("%d", s.store.todoIndex),
		Title:     req.Title,
		Completed: false,
	}
	s.store.Add(todo)
	s.store.todoIndex++
	s.broker.Publish(todo)
	return todo, nil
}

func (s *todoServiceServer) GetTodos(ctx context.Context, req *emptypb.Empty) (*todopb.GetTodosResponse, error) {
	todos := s.store.GetAll()
	return &todopb.GetTodosResponse{Todos: todos}, nil
}

func (s *todoServiceServer) UpdateTodo(ctx context.Context, req *todopb.UpdateTodoRequest) (*todopb.Todo, error) {
	newTodoTitle := s.store.Get(req.Id).Title
	newTodoCompleted := s.store.Get(req.Id).Completed

	if req.Title != nil {
		newTodoTitle = *req.Title
	}
	if req.Completed != nil {
		newTodoCompleted = *req.Completed
	}

	todo := &todopb.Todo{
		Id:        req.Id,
		Title:     newTodoTitle,
		Completed: newTodoCompleted,
	}

	s.store.Update(todo)
	s.broker.Publish(todo)
	return todo, nil
}

func (s *todoServiceServer) DeleteTodo(ctx context.Context, req *todopb.DeleteTodoRequest) (*emptypb.Empty, error) {
	s.store.Delete(req.Id)
	s.broker.Publish(&todopb.Todo{Id: req.Id})
	return &emptypb.Empty{}, nil
}

func (s *todoServiceServer) WatchTodos(stream grpc.BidiStreamingServer[todopb.WatchTodosRequest, todopb.Todo]) error {
	log.Println("WatchTodos stream started for a new client")

	var id string
	var ch chan *todopb.Todo
	var watching bool

	reqCh := make(chan *todopb.WatchTodosRequest)

	go func() {
		for {
			req, err := stream.Recv()
			if err != nil {
				close(reqCh)
				return
			}
			reqCh <- req
		}
	}()

	for {
		select {
		case req, ok := <-reqCh:
			if !ok {
				return nil
			}
			switch req.GetAction().(type) {
			case *todopb.WatchTodosRequest_Start:
				log.Println("Start watching todos")
				id, ch = s.broker.Subscribe()
				watching = true
			case *todopb.WatchTodosRequest_Stop:
				s.broker.Unsubscribe(id)
				watching = false
				log.Println("Stopped watching todos")
				return nil
			}
		case todo := <-ch:
			if !watching{
				return nil
			}
			if err := stream.Send(todo); err != nil {
				return err
			}
		case <-stream.Context().Done():
			if watching {
				s.broker.Unsubscribe(id)
			}
			return nil
		}
	}
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

	store := NewTodoStore()
	broker := NewBroker()

	server := &todoServiceServer{
		store:  store,
		broker: broker,
	}

	todopb.RegisterTodoServiceServer(s, server)

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
