.PHONY: proto-gen proto-clean test testfunc

# Generate protobuf files
proto-gen:
	@echo "Generating protobuf files..."
	@mkdir -p gen/proto
	protoc --go_out=gen/proto --go_opt=paths=source_relative \
		--go-grpc_out=gen/proto --go-grpc_opt=paths=source_relative \
		proto/schema/*.proto proto/api/*.proto
	@echo "Protobuf files generated in gen/proto"
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

run:
	@echo "Starting the application..."
	TB_NO_IO_URING=1 go run cmd/grpc/main.go
testfunc:
	@unbuffer richgo test ./... -run $(FUNC) -v \
	| grep -v "no test" | grep -v "^?"

test:
	@unbuffer richgo test ./... -v 2>&1 | grep -v "no test" | grep -v "^?"

# Watch specific test when files change
watchfunc:
	@echo "Watching for changes to run test $(FUNC)..."
	@reflex -r '\.go$$' -- sh -c "unbuffer richgo test ./... -run $(FUNC) -v \
	| grep -v 'no test' | grep -v '^?'"
# Watch all tests when files change
watch:
	@echo "Watching for changes to run all tests..."
	@reflex -r '\.go$$' -- sh -c "unbuffer richgo test ./... -v \
	| grep -v 'no test' | grep -v '^?'"
