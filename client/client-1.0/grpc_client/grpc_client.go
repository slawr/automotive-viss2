/**
* (C) 2023 Ford Motor Company
*
* All files and artifacts in the repository at https://github.com/w3c/automotive-viss2
* are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/
package main

import (
	"fmt"
	"context"
	"time"
	"os"
	"github.com/akamensky/argparse"
	utils "github.com/w3c/automotive-viss2/utils"
	pb "github.com/w3c/automotive-viss2/grpc_pb"
	"google.golang.org/grpc"
)

const (
    address     = "localhost:5000"
    name = "VISSv2-gRPC-client"
)

var grpcCompression utils.Compression

var commandList []string

func initCommandList() {
	commandList = make([]string, 4)
	commandList[0] = `{"action":"get","path":"Vehicle/Cabin/Door/Row1/Right/IsOpen","requestId":"232"}`
	commandList[1] = `{"action":"subscribe","path":"Vehicle/Cabin/Door/Row1/Right/IsOpen","filter":{"type":"timebased","parameter":{"period":"3000"}},"requestId":"246"}`
	commandList[2] = `{"action":"unsubscribe","subscriptionId":"1","requestId":"240"}`
	commandList[3] = `{"action":"set", "path":"Vehicle/Cabin/Door/Row1/Right/IsOpen", "value":"999", "requestId":"245"}`
}

func noStreamCall(commandIndex int) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Printf("did not connect: %v", err)
		return
	}
	defer conn.Close()
	client := pb.NewVISSv2Client(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	vssRequest := commandList[commandIndex]
	var vssResponse string
	if commandIndex == 0 {
		pbRequest := utils.GetRequestJsonToPb(vssRequest, grpcCompression)
		pbResponse, _ := client.GetRequest(ctx, pbRequest)
		vssResponse = utils.GetResponsePbToJson(pbResponse, grpcCompression)
	} else if commandIndex == 2 {
		pbRequest := utils.UnsubscribeRequestJsonToPb(vssRequest, grpcCompression)
		pbResponse, _ := client.UnsubscribeRequest(ctx, pbRequest)
		vssResponse = utils.UnsubscribeResponsePbToJson(pbResponse, grpcCompression)
	} else {
		pbRequest := utils.SetRequestJsonToPb(vssRequest, grpcCompression)
		pbResponse, _ := client.SetRequest(ctx, pbRequest)
		vssResponse = utils.SetResponsePbToJson(pbResponse, grpcCompression)
	}
	if err != nil {
		fmt.Printf("Error when issuing request=:%s", vssRequest)
	} else {
		fmt.Printf("Received response:%s", vssResponse)
	}
}

func streamCall(commandIndex int) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Printf("did not connect: %v", err)
		return
	}
	defer conn.Close()
	client := pb.NewVISSv2Client(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vssRequest := commandList[commandIndex]
	pbRequest := utils.SubscribeRequestJsonToPb(vssRequest, grpcCompression)
	stream, err := client.SubscribeRequest(ctx, pbRequest)
	for {
		pbResponse, err := stream.Recv()
		vssResponse := utils.SubscribeStreamPbToJson(pbResponse, grpcCompression)
		if err != nil {
			fmt.Printf("Error=%v when issuing request=:%s", err, vssRequest)
		} else {
			fmt.Printf("Received response:%s\n", vssResponse)
		}
	}
}

func main() {
	// Create new parser object
	parser := argparse.NewParser("print", "gRPC client")
	// Create string flag
	logFile := parser.Flag("", "logfile", &argparse.Options{Required: false, Help: "outputs to logfile in ./logs folder"})
	logLevel := parser.Selector("", "loglevel", []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}, &argparse.Options{
		Required: false,
		Help:     "changes log output level",
		Default:  "info"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	utils.InitLog("grpc_client-log.txt", "./logs", *logFile, *logLevel)
	grpcCompression = utils.PB_LEVEL1
	initCommandList()

	fmt.Printf("Command indicies: 0=GET, 1=SUBSCRIBE, 2=UNSUBSCRIBE, 3=SET, any other value terminates.\n")
	var commandIndex int
	for {
		fmt.Printf("\nCommand index [0-3]:")
		fmt.Scanf("%d", &commandIndex)
		if (commandIndex < 0 || commandIndex > 3) {
			break
		}
		fmt.Printf("Command:%s", commandList[commandIndex])
		if (commandIndex == 1) { // subscribe
			go streamCall(commandIndex)
		} else {
			go noStreamCall(commandIndex)
		}
		time.Sleep(1 * time.Second)
	}
}
