VERSION=1.0.0

default:
	GO111MODULE=on go build -o bin/mysql-event ./cmd/mysql-event.go
