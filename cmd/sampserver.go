/*
Sample collection server.

Recieves connection from clients, streams current samples.
*/
package main

import (
	"fmt"
	"flag"
	"net"
	"log"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	//"google.golang.org/protobuf/proto"

	pb "github.com/mikew686/rpidemo/internal/streamer"
)

var (
	redisaddr = flag.String("redisaddr", "localhost:6379", "address of the redis server in hostname:port format")
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	port       = flag.Int("port", 8118, "The server port")
)

type streamSamplesServer struct {
	pb.UnimplementedStreamSamplesServer
}

func (s *streamSamplesServer) SendSample(stream pb.StreamSamples_SendSampleServer) error {
	var sampleCount int32
	for {
		samp, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.SampleResponse{
				Status: "OK",
				Msg: fmt.Sprintf("recieved", sampleCount, "samples"),
			})
		}
		if err != nil {
			return err
		}
		fmt.Printf("Got sample: %v", samp)
		sampleCount++
	}
}

func newServer() *streamSamplesServer {
	s := &streamSamplesServer{}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *tls {
		/*if *certFile == "" {
			*certFile = data.Path("x509/server_cert.pem")
		}
		if *keyFile == "" {
			*keyFile = data.Path("x509/server_key.pem")
		}*/
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterStreamSamplesServer(grpcServer, newServer())
	fmt.Println("Serving on", *port)
	grpcServer.Serve(lis)
}