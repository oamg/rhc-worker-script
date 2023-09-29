package main

import (
	"os"

	"git.sr.ht/~spc/go-log"

	"github.com/oamg/rhc-worker-script/src/internal"
	"github.com/oamg/rhc-worker-script/src/protocol"
)

// Initialized in main
const configFilePath = "/etc/rhc/workers/rhc-worker-script.yml"
const logDir = "/var/log/rhc-worker-script"
const logFileName = "rhc-worker-script.log"
var config *Config
var yggSocketAddrExists bool // Has to be separately declared otherwise grpc.Dial doesn't work
var yggdDispatchSocketAddr string

// main is the entry point of the application. It initializes values from the
// environment, sets up the logger, establishes a connection with the
// dispatcher, registers as a handler, listens for incoming messages, and
// starts accepting connections as a Worker service.
// Note: The function blocks and runs indefinitely until the server is stopped.
func main() {
	yggdDispatchSocketAddr , yggSocketAddrExists = os.LookupEnv("YGG_SOCKET_ADDR")
	if !yggSocketAddrExists {
		log.Fatal("Missing YGG_SOCKET_ADDR environment variable")
	}
	logFile := internal.SetupLogger(logDir, logFileName)
	defer logFile.Close()

	config = loadConfigOrDefault(configFilePath)
	defer os.Remove(*config.TemporaryWorkerDirectory)

	protocol.StartRoutine()
}
