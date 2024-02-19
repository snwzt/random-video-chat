build:
	@go build -o bin/rvc-user cmd/user/main.go
	@go build -o bin/rvc-chat cmd/chat/main.go
	@go build -o bin/rvc-forwarder cmd/forwarder/main.go

run-user:
	@./bin/rvc-user

run-chat:
	@./bin/rvc-chat

run-forwarder:
	@./bin/rvc-forwarder

clean:
	@rm -rf bin