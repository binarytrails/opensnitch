package api

/*
Package api holds the functionality to interact with opensnitch
nodes/clients.

If a new client wants to connect to the server (UI, proxy2db, ...),
it must follow these steps:

	1. Subscribe() - tell the server who we are.
	2. Notifications() - open and keep opened a communication channel
	3. Ping() - ping the server every n seconds, and send the statistics.
	4. AskRule() - called when a new outgoing connection is about to be established.

*/

import (
	"net"
	"os"
	"time"

	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/core"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct {
	srv       *protocol.UIServer
	apiClient *Client
}

// Ping receives every second the stats of a node.
func (s *server) Ping(ctx context.Context, ping *protocol.PingRequest) (*protocol.PingReply, error) {
	s.apiClient.UpdateStats(ctx, ping.Stats)
	return &protocol.PingReply{Id: ping.Id}, nil
}

// AskRule waits for action on a new outgoing connection.
// If it not answered, after n seconds it'll apply the default action
func (s *server) AskRule(ctx context.Context, con *protocol.Connection) (*protocol.Rule, error) {
	resultChan := s.apiClient.AskRule(ctx, con)
	select {
	case rule := <-resultChan:
		return rule, nil
		// XXX: the daemon as of v1.0.1 has this timeout hardcoded
	case <-time.After(120 * time.Second):
		// TODO: apply default action
		return nil, nil
	}
}

// Subscribe receives connections from new nodes with their configuration.
// The nodes are saved to keep a list of connected nodes.
func (s *server) Subscribe(ctx context.Context, clientConf *protocol.ClientConfig) (*protocol.ClientConfig, error) {
	s.apiClient.AddNewNode(ctx, clientConf)
	return &protocol.ClientConfig{}, nil
}

// Notifications opens a permanent channel to send commands back to the nodes.
// This function can't return until the connection with the node is closed,
// in order to maintain the communication channel opened.
func (s *server) Notifications(streamChannel protocol.UI_NotificationsServer) error {
	s.apiClient.OpenChannelWithNode(streamChannel)
	return nil
}

// socketIsReady determines if we can listen for new nodes
func socketIsReady(proto, port string) bool {
	if proto != "unix" {
		return true
	}
	// connect to the unix socket:
	// - if we can connect, there's another UI listening for nodes, so we can't continue.
	// - if not, remove the unix socket file to avoid "address already in use" error.
	_, err := net.Dial(proto, port)
	if err == nil {
		log.Fatal("There's another GUI/TUI/*UI running. Please, close it before launching this UI.")
		return false
	}
	if core.Exists(port) {
		if err := os.Remove(port); err != nil {
			log.Fatal("Can not listen for new clients. Try again removing %s first.", port)
		}
	}
	return true
}

// StartServer start listening for incoming nodes/clients.
func startServer(client *Client, proto, port string) {
	if socketIsReady(proto, port) == false {
		return
	}

	sockFd, err := net.Listen(proto, port)
	if err != nil {
		log.Error("failed to listen on %s: %v", port, err)
	}
	// create a server instance
	s := server{}
	s.apiClient = client
	grpcServer := grpc.NewServer()
	protocol.RegisterUIServer(grpcServer, &s)
	// start the server
	if err := grpcServer.Serve(sockFd); err != nil {
		log.Error("failed to listen for new nodes: %s", err)
	}
}
