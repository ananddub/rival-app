.PHONY: proto-gen sqlc-gen gen-all proto-descriptor docker-up docker-down envoy-up envoy-down run

proto-gen:
	protoc --go_out=gen/proto --go_opt=paths=source_relative \
		--go-grpc_out=gen/proto --go-grpc_opt=paths=source_relative \
		proto/schema/*.proto proto/api/*.proto

sqlc-gen:
	sqlc generate

proto-descriptor:
	protoc --include_imports --include_source_info \
		--descriptor_set_out=proto.pb \
		proto/schema/*.proto proto/api/*.proto

gen-all: proto-gen sqlc-gen proto-descriptor

envoy-up: proto-descriptor
	docker compose up -d

envoy-down:
	docker compose down

docker-up: envoy-up

docker-down: envoy-down

run:
	go run cmd/grpc/main.go
