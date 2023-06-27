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
		fileName := writeFileToTemporaryDir(d.GetContent())
		defer os.Remove(fileName)

		log.Infoln("Executing bash script located at: ", fileName)
		// Execute the script we wrote to the file
		executeScript(fileName)

		// TODO(r0x0d): Remove this after PoC. We will be reading the output
		// that comes from the executeScript function, as we want what is in
		// the stdout to make it more generic, instead of relying on reading an
		// output file in the system. https://issues.redhat.com/browse/HMS-2005
		reportFile := "/var/log/convert2rhel/convert2rhel-report.json"
		log.Infoln("Reading output file at: ", reportFile)
		fileContent, boundary := readOutputFile(reportFile)

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
		// Call "Send"
		var data *pb.Data
		log.Infof("Sending message to %s", d.GetMessageId())
		if fileContent != nil {
			contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
			log.Infof("Sending message to %s", d.GetMessageId())
			data = &pb.Data{
				MessageId:  uuid.New().String(),
				ResponseTo: d.GetMessageId(),
				Metadata: map[string]string{
					"Content-Type": contentType,
				},
				Content:   fileContent.Bytes(),
				Directive: d.GetMetadata()["return_url"],
			}
		} else {
			data = &pb.Data{
				MessageId:  uuid.New().String(),
				ResponseTo: d.GetMessageId(),
				Metadata:   d.GetMetadata(),
				Directive:  d.GetDirective(),
			}
		}

		log.Infoln("pb.Data message: ", data)
		if _, err := c.Send(ctx, data); err != nil {
			log.Error(err)
		}
	}()

	// Respond to the start request that the work was accepted.
	return &pb.Receipt{}, nil
}
