//
// client.go
//
// gRPC Push Notification Client
//

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"

	pb "grpc-push-notif/protos"

	"google.golang.org/grpc"
)

var name string
var subChan chan pb.TopicType

const (
	address = "localhost:50051"
)

// register ...
func register(client pb.PushNotifClient) error {
	//log.Printf("Calling Register RPC")

	resp, err := client.Register(context.Background(), &pb.RegistrationRequest{ClientName: name})
	if err != nil {
		log.Fatalf("Register failed %v", err)
	}
    log.Println("Registration Resp: ", resp, err)
	// cookie = resp.ClientCookie
	// log.Printf("Reg Response %s, %s, %v", resp.ClientName, resp.ServerName, resp.ClientCookie)

	return nil
}

// subscribeToAlerts ...
func subscribeToAlerts(client pb.PushNotifClient) {
	log.Println("Subscribing to Alerts")

	stream, err := client.Alert(context.Background(), &pb.Topic{ClientName: name, Type: pb.TopicType_ALERT})
	if err != nil {
		log.Fatalf("failed to subscribe to Alerts %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed to recv alerts")
		}
		log.Print(resp, err)
	}
}

// subscribeToModeChanges ...
func subscribeToModeChanges(c chan pb.TopicType) {
	log.Println("Subscribing to Mode changes")
	c <- pb.TopicType_MODE
}

// subscribeToModeChanges ...
func subscribeToTempChanges(c chan pb.TopicType) {
	log.Println("Subscribing to Temp changes")
	c <- pb.TopicType_TEMP
}

// recvNotification ...
func recvNotification(stream pb.PushNotif_SubscribeClient) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Fatalf("failed to recv %v", err)
		}
		if err == io.EOF {
			break
		}
		log.Print(resp, err)
	}
}

// subscribe ...
func subscribe(client pb.PushNotifClient, c chan pb.TopicType) {
	stream, err := client.Subscribe(context.Background())
	if err != nil {
		log.Fatalf("failed to subscribe %v", err)
	}

	go recvNotification(stream)

	for t := range c {
		//log.Printf("sending subscibe for topic %s", t.String())
		if err := stream.Send(&pb.Topic{ClientName: name, Type: t}); err != nil {
			log.Fatalf("couldnt send %s", t.String())
		}
	}
	stream.CloseSend()
}

// main ...
func main() {

	// init important structures
	subChan = make(chan pb.TopicType, 10)
	rand.Seed(time.Now().UTC().UnixNano())
	name = fmt.Sprintf("%s:%d", "Client", rand.Intn(50))

	// Setup a connection with the server
	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := pb.NewPushNotifClient(conn)

	register(client)

	// Subscribe to alerts.. a way to request for push notifications.
	// go routine below keeps the stream long lived and listens to
	// alerts as long as the connection is up
	go subscribeToAlerts(client)

	subscribeToModeChanges(subChan)
	subscribeToTempChanges(subChan)
	subscribe(client, subChan)
}
