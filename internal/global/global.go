package global

import (
	"context"
)

type response struct {
	Message string `json:"message"`
}

//encore:api public method=GET path=/hello
func Hello(ctx context.Context) (response, error) {
	return response{Message: "Hello, World!"}, nil
}
