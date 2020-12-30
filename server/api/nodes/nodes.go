package nodes

import (
	"net"
	"time"

	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
)

type nodeStats struct {
	events []*protocol.Event
	n      *node
}

var (
	nodeList  = make(map[string]*node)
	statsList = make(map[string]*nodeStats)
)

// Add a new node the list of nodes.
func Add(ctx context.Context, nodeConf *protocol.ClientConfig) {
	addr := GetAddr(ctx)
	if addr == "" {
		log.Warning("node not added, invalid addr: %v", GetPeer(ctx))
		return
	}
	nodeList[addr] = NewNode(ctx, addr, nodeConf)
}

// SetNotificationsChannel sets the communication channel for a given node.
// https://github.com/grpc/grpc-go/blob/master/stream.go
func SetNotificationsChannel(notificationsStream protocol.UI_NotificationsServer) *node {
	ctx := notificationsStream.Context()
	addr := GetAddr(ctx)
	// ctx.AddCallback() ?
	if !isConnected(addr) {
		log.Warning("nodes.SetNotificationsChannel() not found: %s", addr)
		return nil
	}
	nodeList[addr].NotificationsStream = notificationsStream

	return nodeList[addr]
}

func SendNotifications(notif *protocol.Notification) {
	for _, node := range nodeList {
		node.SendNotification(notif)
	}
}

// UpdateStats of a node.
func UpdateStats(ctx context.Context, stats *protocol.Statistics) {
	addr := GetAddr(ctx)
	if !isConnected(addr) {
		log.Warning("nodes.UpdateStats() not found: %s", addr)
		return
	}
	nodeList[addr].UpdateStats(stats)
}

// Delete a node from the list of nodes.
func Delete(n *node) bool {
	n.Close()
	delete(nodeList, n.Addr())
	return true
}

// Get a node from the list of nodes.
func Get(addr string) *node {
	return nodeList[addr]
}

// GetPeer gets the address:port of a node.
func GetPeer(ctx context.Context) *peer.Peer {
	p, _ := peer.FromContext(ctx)
	return p
}

// GetAddr of a node from the context
func GetAddr(ctx context.Context) (addr string) {
	p := GetPeer(ctx)
	host, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil && p.Addr.String() == "@" {
		host = "localhost"
		addr = "unix://" + host
	} else if err != nil {
		log.Error("nodes.GetAddr() can not get noe address, addr:", p.Addr.String())
		return ""
	}
	addr = p.Addr.Network() + ":" + host
	return
}

// GetAll nodes.
func GetAll() *map[string]*node {
	return &nodeList
}

// GetStats returns the stats of all nodes combined.
func GetStats() (stats []*protocol.Statistics) {
	for addr, node := range *GetAll() {
		println(addr, node)
	}

	return stats
}

// GetStatsSum returns total connections for a particular view
func GetStatsSum(what int) (cons uint64) {
	for _, node := range *GetAll() {
		if node.GetStats() == nil {
			continue
		}
		switch what {
		case 0:
			cons += node.GetStats().Connections
		case 1:
			cons += node.GetStats().Dropped
		case 2:
			cons += node.GetStats().Rules
		case 3:
			cons += uint64(len(node.GetStats().Events))
		}
	}

	return cons
}

// Total returns the number of active nodes.
func Total() int {
	return len(nodeList)
}

func isConnected(addr string) bool {
	_, found := nodeList[addr]
	return found
}

// StartFirewall starts the interception of connections on the nodes
func StartFirewall() {
	SendNotifications(
		&protocol.Notification{
			Id:   uint64(time.Now().UnixNano()),
			Type: protocol.Action_LOAD_FIREWALL,
		})
}

// StopFirewall stops connections interception on the nodes
func StopFirewall() {
	SendNotifications(
		&protocol.Notification{
			Id:   uint64(time.Now().UnixNano()),
			Type: protocol.Action_UNLOAD_FIREWALL,
		})
}
