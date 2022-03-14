package image_monitor

import (
	"context"
	"fmt"
	"log"
	"omni-manager/util"
	"time"

	grpc "google.golang.org/grpc"
)

type monitorServer struct {
	UnimplementedCallCenterServer
}
type CallCenterService struct {
}

func (centerService *CallCenterService) RegisterService(ctx context.Context, in *ClientRequest) {

}

//start monitor
func StartMonitor(address string) {

	conn, err := grpc.Dial(address) //, grpc.WithInsecure()
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	userClient := NewCallCenterClient(conn)
	// timeout 3s
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// PullFromDispatcher request
	PullFromDispatcherReponse, err := userClient.PullFromDispatcher(ctx, &DispatcherRequest{Name: "dfdf"})
	if err != nil {
		util.Log.Printf("user index could not greet: %v", err)
	}

	if PullFromDispatcherReponse.Name == "name" {
		util.Log.Printf("user index success: %s", PullFromDispatcherReponse.Name)

		fmt.Println(PullFromDispatcherReponse.Name)

	} else {
		util.Log.Printf("user index error: %d", PullFromDispatcherReponse.Name)
	}
}
