package v1

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestDial(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:38849", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}

	client := NewGitserverServiceClient(conn)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := client.Exec(ctx, nil)
		if err != nil {
			cancel()
			fmt.Println(err.Error())
		}
		cancel()
	}
}
