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
const configFilePath = "/etc/rhc/workers/rhc-worker-script.yml"
const logDir = "/var/log/rhc-worker-script"
const logFileName = "rhc-worker-script.log"

var yggdDispatchSocketAddr string
var config *Config

// main is the entry point of the application. It initializes values from the
// environment, sets up the logger, establishes a connection with the
// dispatcher, registers as a handler, listens for incoming messages, and
// starts accepting connections as a Worker service.
// Note: The function blocks and runs indefinitely until the server is stopped.
func main() {
	var yggSocketAddrExists bool // Has to be separately declared otherwise grpc.Dial doesn't work
	yggdDispatchSocketAddr, yggSocketAddrExists = os.LookupEnv("YGG_SOCKET_ADDR")
	if !yggSocketAddrExists {
		log.Fatal("Missing YGG_SOCKET_ADDR environment variable")
	}
	logFile := setupLogger(logDir, logFileName)
	defer logFile.Close()

	config = loadConfigOrDefault(configFilePath)
	defer os.Remove(*config.TemporaryWorkerDirectory)

	// Dial the dispatcher on its well-known address.
	conn, err := grpc.Dial(
		yggdDispatchSocketAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
