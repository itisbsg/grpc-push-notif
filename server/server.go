//
// server.go
//
// gRPC Push Notification Server
//

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	pb "grpc-push-notif/protos"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	name      string
	alertStrm map[string]pb.PushNotif_AlertServer
	subStrm   map[string]pb.PushNotif_SubscribeServer
}

func newServer() *server {
	return &server{name: "pushnotifServer", alertStrm: make(map[string]pb.PushNotif_AlertServer),
		subStrm: make(map[string]pb.PushNotif_SubscribeServer)}
}

func (s *server) Register(ctx context.Context, req *pb.RegistrationRequest) (*pb.RegistrationResponse, error) {
	log.Printf("Received a Registration Request (%s)", req.GetClientName())
	return &pb.RegistrationResponse{ClientName: req.GetClientName(), ServerName: s.name}, nil
}

func (s *server) Alert(topic *pb.Topic, stream pb.PushNotif_AlertServer) error {
	log.Printf("Received a Subscription Request(%s, %s)", topic.GetClientName(), topic.GetType().String())
	s.alertStrm[topic.GetClientName()] = stream
	//log.Printf("Client %s will now recieve Alerts", topic.GetClientName())
    log.Printf("")

	// long lived stream
	for {
	}

}

func (s *server) Subscribe(stream pb.PushNotif_SubscribeServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		log.Printf("Received a Subscription Request (%s, %s)", in.GetClientName(), in.GetType().String())
		_, ok := s.subStrm[in.GetClientName()+":"+in.GetType().String()]
		if !ok {
			s.subStrm[in.GetClientName()+":"+in.GetType().String()] = stream
			// log.Printf("Client %v will now recieve updates about topic %s", in.GetClientName(), in.GetType().String())
		} else {
			log.Printf("Client %s already subscribed to %s", in.GetClientName(), in.GetType().String())
		}
	}
	return nil
}

func (s *server) pushUpdates() {
	var str string
	var i uint32

	time.Sleep(1 * time.Second)

	for {
		//fmt.Println(s.alertStrm)
		for k, v := range s.alertStrm {
			fmt.Printf("Enter new Alert for Client(%v): ", k)
			fmt.Scan(&str)
			if err := v.Send(&pb.Notification{Type: pb.TopicType_ALERT, A: &pb.Alert{Message: str}}); err != nil {
				log.Fatalf("Send failed %v", err)
			}
		}

		//fmt.Println("walk thru the subscribe streams")
		for k, v := range s.subStrm {
			if strings.Contains(k, "MODE") {
				fmt.Printf("Enter new Mode for Client(%v): ", k)
				fmt.Scan(&str)
				if err := v.Send(&pb.Notification{Type: pb.TopicType_MODE, M: &pb.Mode{NewMode: str}}); err != nil {
					log.Fatalf("Send failed %v", err)
				}
			} else {
				fmt.Printf("Enter new Temp for Client(%v): ", k)
				fmt.Scan(&i)
				if err := v.Send(&pb.Notification{Type: pb.TopicType_TEMP, T: &pb.Temp{NewTemp: i}}); err != nil {
					log.Fatalf("Send failed %v", err)
				}
			}
		}
	}
}

// main ...
func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	myServer := newServer()
	pb.RegisterPushNotifServer(s, myServer)

	go myServer.pushUpdates()

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
