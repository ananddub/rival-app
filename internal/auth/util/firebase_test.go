package util

import (
	"context"
	"testing"
)

func TestFirebase(t *testing.T) {
	ctx := context.Background()
	firbase, err := NewFirebaseService(ctx)
	if err != nil || firbase == nil {
		t.Errorf("Expected Firebase service to be initialized, got nil")

	}

}
