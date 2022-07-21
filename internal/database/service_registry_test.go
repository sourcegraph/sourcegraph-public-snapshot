package database

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestServiceRegistry(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	store := db.Services()
	service := "zoekt-indexserver"

	err := store.Renew(ctx, service, "127.0.0.1:1234")
	if err == nil {
		t.Fatal("the store should have complained because we cannot renew before registering")
	}

	err = store.Deregister(ctx, service, "127.0.0.1:1234")
	if err == nil {
		t.Fatal("the store should have complained because we cannot de-register before registering")
	}

	addr, err := netip.ParseAddr("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	id1, err := store.Register(ctx, service, ServiceArgs{
		IP:       addr,
		Port:     1234,
		Hostname: "host1",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Register(ctx, service, ServiceArgs{
		IP:       addr,
		Port:     1235,
		Hostname: "host2",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.Renew(ctx, service, id1)
	if err != nil {
		t.Fatal(err)
	}

	instances, err := store.GetByService(ctx, service)
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 2 {
		t.Fatal("expected 2 instances")
	}

	err = store.Deregister(ctx, service, id1)
	if err != nil {
		t.Fatal(err)
	}

	instances, err = store.GetByService(ctx, service)
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 1 {
		t.Fatal("expected 1 instance")
	}

	if got := instances[0].Hostname; got != "host2" {
		t.Fatalf("want %q, got %q", "host2", got)
	}

	err = store.Invalidate(ctx, 0*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	instances, err = store.GetByService(ctx, service)
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 0 {
		t.Fatalf("instances: want=0, have=%d", len(instances))
	}
}
