package repository

import (
	"context"
	"errors"
	"testing"
)

func TestListContextHonorsCancellation(t *testing.T) {
	repo := newTestRepository(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := repo.ListContext(ctx, ListFilter{})
	if err == nil {
		t.Fatal("expected list to honor the canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
