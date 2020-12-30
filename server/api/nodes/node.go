package nodes

import (
	"fmt"
	"sync"
	"time"

	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"golang.org/x/net/context"
)

// Status represents the current connectivity status of a node.
type Status string

// Statuses of a node.
var (
	Online  = Status(log.Bold(log.Green("online")))
	Offline = Status(log.Bold(log.Red("offline")))
)

type node struct {
	sync.RWMutex
	// proto:host
	addr                 string
	ctx                  context.Context
	lastSeen             time.Time
	status               Status
	NotificationsStream  protocol.UI_NotificationsServer
	notificationsChannel chan *protocol.Notification
	config               *protocol.ClientConfig
	stats                *protocol.Statistics
}

// NewNode instanstiates a new node.
func NewNode(ctx context.Context, addr string, nodeConf *protocol.ClientConfig) *node {
	log.Info("NewNode: %s - %s, %v", nodeConf.Name, nodeConf.Version, addr)
	return &node{
		addr:                 addr,
		ctx:                  ctx,
		lastSeen:             time.Now(),
		status:               Online,
		config:               nodeConf,
		notificationsChannel: make(chan *protocol.Notification, 1),
	}
}

func (n *node) String() string {
	n.RLock()
	defer n.RUnlock()

	return fmt.Sprintf("[%v]  -  [%-20s]  -  [%-24s]  -  [%s]  -  [%s]", n.lastSeen.Format(time.Stamp), n.addr, n.status, n.config.Version, n.config.Name)
}

// Addr returns the address of the node.
func (n *node) Addr() string {
	return n.addr
}

func (n *node) Close() {
	n.status = Offline
}

func (n *node) Status() Status {
	return n.status
}

// LastSeen returns the last time the node was seen by the server.
func (n *node) LastSeen() time.Time {
	return n.lastSeen
}

// SendNotification to the node via the channel and grpc.ServerStream channel.
func (n *node) SendNotification(notif *protocol.Notification) {
	n.notificationsChannel <- notif
}

func (n *node) UpdateStats(stats *protocol.Statistics) {
	n.Lock()
	defer n.Unlock()

	n.stats = stats
	n.lastSeen = time.Now()
}

func (n *node) GetStats() *protocol.Statistics {
	n.Lock()
	defer n.Unlock()

	return n.stats
}

func (n *node) GetConfig() *protocol.ClientConfig {
	n.RLock()
	defer n.RUnlock()

	return n.config
}

// GetNotifications returns the notifications channel.
func (n *node) GetNotifications() chan *protocol.Notification {
	return n.notificationsChannel
}
