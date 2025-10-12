.PHONY: proto proto-install swagger wire wire-install clean help

# Help target
help:
	@echo "Available targets:"
	@echo "  proto-install  - Install protoc and Go plugins"
	@echo "  proto          - Generate Go code from proto files"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  wire-install   - Install Wire dependency injection tool"
	@echo "  wire           - Generate dependency injection code with Wire"
	@echo "  clean          - Clean generated files"
	@echo "  docker-up      - Start docker containers"
	@echo "  docker-down    - Stop docker containers"
	@echo "  run-user       - Run user service locally"
	@echo "  run-product    - Run product service locally"

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

# Install Wire
wire-install:
	@echo "Installing Wire dependency injection tool..."
	go install github.com/google/wire/cmd/wire@latest
	@echo "Wire installation complete!"

# Generate dependency injection code with Wire
wire:
	@echo "Generating dependency injection code with Wire..."
	@cd internal/user && go generate
	@cd internal/product && go generate
	@echo "Wire generation complete!"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f api/proto/user/*.pb.go
	rm -f internal/user/wire_gen.go
	rm -f internal/product/wire_gen.go

# Docker commands
docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

# Run services locally
run-user:
	go run cmd/user/main.go

run-product:
	go run cmd/product/main.go

