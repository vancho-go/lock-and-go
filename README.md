# lock-and-go
Client-server system that allows the user to reliably and securely store logins, passwords, binary data and other private information.

## Начало работы

### Предварительные требования

Запуск сервера:

```golang
go run cmd/server/main.go -a :8080 -l debug -js "test-key-secret" -jt 24h -d "your_db_dsn"
