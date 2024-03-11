/*
Sample collection client, to be run on the Rasberry Pi.

Connects to redis, loads current samples, and streams them to the sample server.
*/
package main

import (
	"fmt"
	"flag"
	"log"
	"time"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/mikew686/rpidemo/internal/streamer"
)

var (
	redisaddr = flag.String("redisaddr", "localhost:6379", "address of the redis server in hostname:port format")
	servaddr = flag.String("servaddr", "localhost:8118", "address of the sample server in hostname:port format")
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverHostOverride = flag.String("server_host_override", "rpitest.test.com", "The server name used to verify the hostname returned by the TLS handshake")
)

func testSendSamples(client pb.StreamSamplesClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.SendSample(ctx)
	if err != nil {
		log.Fatalf("client.SendSample failed: %v", err)
	}
	samp := &pb.SampleRequest{
		Timestamp: 1234,
		LightLevelPercent: 75.0,
		TemperatureCelcius: 18.2,
		HumidityPercent: 50.1,
	}
	log.Printf("Sending:", samp)
	if err := stream.Send(samp); err != nil {
		log.Fatalf("client.SendSample: stream.Send(%v) failed: %v", samp, err)
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("client.SendSample failed: %v", err)
	}
	log.Printf("Reply: %v", reply)
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		/*if *caFile == "" {
			*caFile = data.Path("x509/ca_cert.pem")
		}*/
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(*servaddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected to", *servaddr)
	client := pb.NewStreamSamplesClient(conn)
	testSendSamples(client)
}