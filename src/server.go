package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"git.sr.ht/~spc/go-log"
	"github.com/google/uuid"
	pb "github.com/redhatinsights/yggdrasil/protocol"
	"google.golang.org/grpc"
)

// jobServer implements the Worker gRPC service as defined by the yggdrasil
// gRPC protocol. It accepts Assignment messages, unmarshals the data into a
// string, and echoes the content back to the Dispatch service by calling the
// "Finish" method.
type jobServer struct {
	pb.UnimplementedWorkerServer
}

// Send implements the "Send" method of the Worker gRPC service.
func (s *jobServer) Send(ctx context.Context, d *pb.Data) (*pb.Receipt, error) {
	go func() {
		log.Infoln("Writing temporary bash script.")
		// Write the file contents to the temporary disk
		scriptFileName := writeFileToTemporaryDir(d.GetContent())
		defer os.Remove(scriptFileName)

		log.Infoln("Executing and reading output of bash script located at: ", scriptFileName)
		// Execute the script we wrote to the file
		commandOutput := executeScript(scriptFileName)

		// Dial the Dispatcher and call "Finish"
		conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithInsecure())
		if err != nil {
			log.Error(err)
		}
		defer conn.Close()

		// Create a client of the Dispatch service
		c := pb.NewDispatcherClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// Create a data message to send back to the dispatcher.
		log.Infof("Creating payload for message %s", d.GetMessageId())

		correlationID := d.GetMetadata()["correlation_id"]
		metadataContentType := d.GetMetadata()["return_content_type"]
		fileContent, boundary := getOutputFile(scriptFileName, commandOutput, correlationID, metadataContentType)

		var data *pb.Data
		if commandOutput != "" && fileContent != nil {
			contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
			log.Infof("Sending message to %s", d.GetMessageId())
			data = &pb.Data{
				MessageId:  uuid.New().String(),
				ResponseTo: d.GetMessageId(),
				Metadata:   constructMetadata(d.GetMetadata(), contentType),
				Content:    fileContent.Bytes(),
				Directive:  d.GetMetadata()["return_url"],
			}
		} else {
			data = &pb.Data{
				MessageId:  uuid.New().String(),
				ResponseTo: d.GetMessageId(),
				Metadata:   d.GetMetadata(),
				Directive:  d.GetDirective(),
			}
		}

		// Call "Send"
		log.Infof("Sending message to %s", d.GetMessageId())
		log.Infoln("pb.Data message: ", data)
		if _, err := c.Send(ctx, data); err != nil {
			log.Error(err)
		}
	}()

	// Respond to the start request that the work was accepted.
	return &pb.Receipt{}, nil
}
