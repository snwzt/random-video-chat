build:
	@go build -o bin/rvc cmd/main.go

run-user:
	@./bin/rvc user-service

run-chat:
	@./bin/rvc chat-service

run-forwarder:
	@./bin/rvc forwarder-service

clean:
	@rm -rf bin