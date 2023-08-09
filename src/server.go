package main

import (
	"context"
	"fmt"
	"time"

	"git.sr.ht/~spc/go-log"
	"github.com/google/uuid"
	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Create message payload
func createDataMessage(commandOutput string, metadata map[string]string, directive string, messageID string) *pb.Data {
	correlationID := metadata["correlation_id"]
	metadataContentType := metadata["return_content_type"]
	fileContent, boundary := getOutputFile(commandOutput, correlationID, metadataContentType)

	var data *pb.Data
	if commandOutput != "" && fileContent != nil {
		contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
		log.Infof("Sending message to %s", messageID)
		data = &pb.Data{
			MessageId:  uuid.New().String(),
			ResponseTo: messageID,
			Metadata:   constructMetadata(metadata, contentType),
			Content:    fileContent.Bytes(),
			Directive:  metadata["return_url"],
		}
	} else {
		data = &pb.Data{
			MessageId:  uuid.New().String(),
			ResponseTo: messageID,
			Metadata:   metadata,
			Directive:  directive,
		}
	}
	return data
}

// Processes signed script and sends message back to dispatcher
func processData(d *pb.Data) *pb.Data {
	log.Infoln("Processing received yaml data")
	commandOutput := processSignedScript(d.GetContent())

	// Create a data message to send back to the dispatcher.
	log.Infof("Creating payload for message %s", d.GetMessageId())
	data := createDataMessage(commandOutput, d.GetMetadata(), d.GetDirective(), d.GetMessageId())
	return data
}

// Sends data back to dispatcher
func sendDataToDispatcher(data *pb.Data) *pb.Data {
	// Dial the Dispatcher and call "Finish"
	conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err)
	}
	defer conn.Close()

	// Create a client of the Dispatch service
	client := pb.NewDispatcherClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Call "Send"
	log.Infof("Sending response message to %s", data.GetResponseTo())
	log.Infoln("pb.Data message: ", data)
	if _, err := client.Send(ctx, data); err != nil {
		log.Error(err)
	}
	return data
}

// jobServer implements the Worker gRPC service as defined by the yggdrasil
// gRPC protocol. It accepts Assignment messages, unmarshals the data into a
// string, and echoes the content back to the Dispatch service by calling the
// "Finish" method.
type jobServer struct {
	pb.UnimplementedWorkerServer
}

// Send is the implementation of the "Send" method of the Worker gRPC service.
// It executes a temporary script, reads its output, and sends a message
// containing the script's result to the Dispatcher service.
//
// The function performs the following steps:
//  1. Writes the contents of the received data to a temporary file on disk.
//  2. Executes the script by calling the appropriate function.
//  3. Establishes a connection with the Dispatcher service using gRPC.
//  4. Creates a client of the Dispatcher service.
//  5. Constructs a data message to send back to the dispatcher.
//  6. Sends the data message using the "Send" method of the Dispatcher service.
func (s *jobServer) Send(_ context.Context, d *pb.Data) (*pb.Receipt, error) {

	// Goroutine processing the data, cancels the context when processing is done
	go func() {
		data := processData(d)
		sendDataToDispatcher(data)
	}()

	// Respond to the start request that the work was accepted.
	return &pb.Receipt{}, nil
}
