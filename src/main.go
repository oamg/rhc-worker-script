package main

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

// Initialized in main
var yggdDispatchSocketAddr string
var logFolder string
var logFileName string
var temporaryWorkerDirectory string
var shouldDoInsightsCoreGPGCheck string
var shouldVerifyYaml string

// main is the entry point of the application. It initializes values from the environment,
// sets up the logger, establishes a connection with the dispatcher, registers as a handler,
// listens for incoming messages, and starts accepting connections as a Worker service.
// Note: The function blocks and runs indefinitely until the server is stopped.
func main() {
	initializedOK, errorMsg := initializeEnvironment()
	if errorMsg != "" && !initializedOK {
		log.Fatal(errorMsg)
	}

	logFile := setupLogger(logFolder, logFileName)
	defer logFile.Close()

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

	// Register as a handler of the "rhc-worker-bash" type.
	r, err := c.Register(ctx, &pb.RegistrationRequest{Handler: "rhc-worker-bash", Pid: int64(os.Getpid()), DetachedContent: true})
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

	log.Infoln("Listening to messages...", yggdDispatchSocketAddr)

	// Register as a Worker service with gRPC and start accepting connections.
	s := grpc.NewServer()
	pb.RegisterWorkerServer(s, &jobServer{})
	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}
}
