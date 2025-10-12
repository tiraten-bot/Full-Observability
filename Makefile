.PHONY: proto proto-install swagger clean help

# Help target
help:
	@echo "Available targets:"
	@echo "  proto-install  - Install protoc and Go plugins"
	@echo "  proto          - Generate Go code from proto files"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  clean          - Clean generated files"
	@echo "  docker-up      - Start docker containers"
	@echo "  docker-down    - Stop docker containers"

# Install protoc plugins
proto-install:
	@echo "Installing protoc Go plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p api/proto/user
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/user/user.proto
	@echo "Proto generation complete!"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@which swag || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/user/docs.go -o cmd/user/docs --parseDependency --parseInternal
	@echo "Swagger generation complete!"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f api/proto/user/*.pb.go

# Docker commands
docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

# Run the service locally
run:
	go run cmd/user/main.go

