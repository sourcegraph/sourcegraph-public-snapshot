# Overview 

New internal APIs must have both gRPC and REST implementations to reduce the maintenance burden. The REST implementation should use the c[canonical JSON representation of the generated protobuf structs](https://protobuf.dev/programming-guides/proto3/#json) in the HTTP body for arguments and responses.



## simple example 

The following example demonstrates how to implement a simple service in Go that provides both gRPC and REST APIs, using the  [canonical JSON representation of the generated Protobuf structs](https://protobuf.dev/programming-guides/proto3/#json). 

Note that the Go service uses [google.golang.org/protobuf/encoding/protojson](https://google.golang.org/protobuf/encoding/protojson) to Marshall and Unmarshall Protobuf structs to/from JSON. The standard "encoding/json" package should **not** be used here: it doesn't correctly operate on protobuf structs. 

### gRPC definition 


```proto 
syntax = "proto3";

package greeting;

service GreeterService {
  rpc SayHello (HelloRequest) returns (HelloReply);
}

message HelloRequest {
  string name = 1;
}

message HelloReply {
  string message = 1;
}
```

### generate the Go protobuf structs 

Create the following buf configuration file: 

#### buf.gen.yaml

The buf configuration file is used to generate the Go code for the Protobuf definition. This file specifies the plugins to use and the output directory for the generated code. The generated code includes the Protobuf structs that we can reuse in both gRPC and REST implementations.

```yaml
# Configuration file for https://buf.build/, which we use for Protobuf code generation.
version: v1
plugins:
  - plugin: buf.build/protocolbuffers/go:v1.29.1
    out: .
    opt:
      - paths=source_relative
  - plugin: buf.build/grpc/go:v1.3.0
    out: .
    opt:
      - paths=source_relative
```

Now, run `sg generate buf` to use the above configuration file to generate the Go code for the protobuf defintion above. This will create the following files:

#### greeter.pb.go

```go
package greeter

type HelloRequest struct { 
	// ...
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

type HelloReply struct {
	// ...
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

// ... (omitted)
```

##### greeter_grpc.pb.go
```go 
package greeter

import (
	context "context"
	grpc "google.golang.org/grpc"
)

type GreeterServiceClient interface {
	SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error)
}

// ... (omitted)
```

### go service implementation 

```go
package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"example.com/greeting"

	// 🚨🚨🚨 note the use of this package instead of "encoding/json" - encoding/json doesn't correctly serialize protobuf structs 
	"google.golang.org/protobuf/encoding/protojson" 
)

type server struct {
	greeting.UnimplementedGreeterServiceServer
}

func (s *server) SayHello(ctx context.Context, in *greeting.HelloRequest) (*greeting.HelloReply, error) {
	return &greeting.HelloReply{Message: "Hello, " + in.Name}, nil
}

func main() {
	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	greeting.RegisterGreeterServiceServer(grpcServer, &server{})
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start REST server
	r := mux.NewRouter()
	r.HandleFunc("/sayhello", sayHelloREST).Methods("POST")
	http.ListenAndServe(":8080", r)
}

func sayHelloREST(w http.ResponseWriter, r *http.Request) {
	var req greeting.HelloRequest
	err := protojson.Unmarshal(r.Body, &req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	resp := &greeting.HelloReply{Message: "Hello, " + req.Name}
	jsonBytes, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
```
