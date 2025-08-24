# grpc-practice

## 使い方

### サーバー起動
```bash
go run cmd/server/main.go
```

### API呼び出し

#### grpcurl使用
```bash
# CreateTodo
grpcurl -plaintext -d '{"title": "新しいタスク"}' localhost:8080 todo.TodoService/CreateTodo

# GetTodos
grpcurl -plaintext -d '{}' localhost:8080 todo.TodoService/GetTodos

# UpdateTodo
grpcurl -plaintext -d '{"id": "1", "title": "更新されたタスク", "completed": true}' localhost:8080 todo.TodoService/UpdateTodo

# DeleteTodo
grpcurl -plaintext -d '{"id": "1"}' localhost:8080 todo.TodoService/DeleteTodo

# サービス一覧
grpcurl -plaintext localhost:8080 list

# メソッド一覧
grpcurl -plaintext localhost:8080 list todo.TodoService
```

#### buf curl使用
```bash
# CreateTodo
buf curl --protocol grpc -d '{"title": "新しいタスク"}' --http2-prior-knowledge http://localhost:8080/todo.TodoService/CreateTodo
```