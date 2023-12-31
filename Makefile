all_health: fmt vet test test_race
health_no_race: fmt vet test

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -parallel 4 -v ./...

test_race:
	go test ./... -short -race

build_examples:
	go build -o ./bin/client ./examples/client/client.go
	go build -o ./bin/rclient ./examples/retryable_client/client.go
	go build -o ./bin/server ./examples/server.go
	go build -o ./bin/rserver ./examples/retryable_server/server.go
	go build -o ./bin/tlsserver ./examples/tls_server/server.go
	go build -o ./bin/tlsclient ./examples/tls_client/client.go