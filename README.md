# микросервис Api-gateway для сервиса Eventify

Команда для генерации из .proto файлово (.proto файлы должны лежать в папке /proto)

```bash
protoc -I proto --go_out=. --go-grpc_out=. proto/auth.proto proto/event.proto
```