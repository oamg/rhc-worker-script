package main

import (
	"context"
	"net"
	"os"
	"time"

	"git.sr.ht/~spc/go-log"

	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
)

// FIXME: Set all contants from ENV variables
var yggdDispatchSocketAddr string
var logFolder string
var logFileName string
var temporaryWorkerDirectory string

func main() {
	// Get initialization values from the environment.
	yggdDispatchSocketAddr, yggSocketAddrExists := os.LookupEnv("YGG_SOCKET_ADDR")
	if !yggSocketAddrExists {
		log.Fatal("Missing YGG_SOCKET_ADDR environment variable")
	}
	logFolder = getEnv("RHC_WORKER_BASH_LOG_FOLDER", "/var/log/rhc-worker-bash")
	logFileName = getEnv("RHC_WORKER_BASH_LOG_FILENAME", "rhc-worker-bash.log")
	temporaryWorkerDirectory = getEnv("RHC_WORKER_BASH_TMP_DIR", "/var/lib/rhc-worker-bash")

	logFile := setupLogger(logFolder, logFileName)
	defer logFile.Close()

	// Dial the dispatcher on its well-known address.
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithInsecure())
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
