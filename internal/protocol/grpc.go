package protocol

import (
	"context"
	"net"
	"os"
	"time"

	"git.sr.ht/~spc/go-log"

	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcProtocolOptions struct {
	grpcSocketAddr string
	pid int64
}

// jobServer implements the Worker gRPC service as defined by the yggdrasil
// gRPC protocol. It accepts Assignment messages, unmarshals the data into a
// string, and echoes the content back to the Dispatch service by calling the
// "Finish" method.
type jobServer struct {
	pb.UnimplementedWorkerServer
}

func StartRoutine(yggdDispatchSocketAddr string) {
	// Dial the dispatcher on its well-known address.
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create a dispatcher client
	c := pb.NewDispatcherClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Register as a handler of the "rhc-worker-script" type.
	r, err := c.Register(
		ctx,
		&pb.RegistrationRequest{
			Handler:         "rhc-worker-script",
			Pid:             int64(os.Getpid()),
			DetachedContent: true,
		})
	if err != nil {
		log.Fatal(err)
	}
	if !r.GetRegistered() {
		log.Fatalf("handler registration failed: %v", err)
	}

	// Listen on the provided socket address.
	l, err := net.Listen("unix", r.GetAddress())
	if err != nil {
		log.Fatal(err)
	}

	// Register as a Worker service with gRPC and start accepting connections.
	s := grpc.NewServer()
	pb.RegisterWorkerServer(s, &jobServer{})
	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}
}

func (options *GrpcProtocolOptions) NewGrpcConnection() (){}
