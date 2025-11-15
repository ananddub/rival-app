.PHONY: proto-gen proto-clean

# Generate protobuf files
proto-gen:
	@echo "Generating protobuf files..."
	@mkdir -p gen/proto
	protoc --go_out=gen/proto --go_opt=paths=source_relative \
		--go-grpc_out=gen/proto --go-grpc_opt=paths=source_relative \
		proto/schema/*.proto proto/api/*.proto

# Clean generated files
proto-clean:
	@echo "Cleaning generated protobuf files..."
	@rm -rf gen/proto

# Install protoc dependencies
proto-deps:
	@echo "Installing protoc dependencies..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Run database migrations
migrate-up:
	@echo "Running database migrations..."
	goose -dir sql/schema postgres "$(DB_URL)" up

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	goose -dir sql/schema postgres "$(DB_URL)" down

# Generate sqlc code
sqlc-gen:
	@echo "Generating sqlc code..."
	sqlc generate

# Run all generations
gen-all: proto-gen sqlc-gen
	@echo "All code generation complete!"
