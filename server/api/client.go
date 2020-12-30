package api

import (
	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"github.com/gustavo-iniguez-goya/opensnitch/daemon/log"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/nodes"
	"github.com/gustavo-iniguez-goya/opensnitch/server/api/storage"
	"golang.org/x/net/context"
	"sync"
)

// Client struct groups the API functionality to communicate with the nodes
type Client struct {
	sync.RWMutex
	db           *storage.Storage
	workAsDaemon bool
	lastStats    *protocol.Statistics
	nodesChan    chan bool
	rulesInChan  chan *protocol.Connection
	rulesOutChan chan *protocol.Rule
}

// rules related constants
const (
	ActionAllow = "allow"
	ActionDeny  = "deny"

	RuleSimple = "simple"
	RuleList   = "list"
	RuleRegexp = "regexp"

	RuleOnce    = "once"
	Rule15s     = "15s"
	Rule30s     = "30s"
	Rule5m      = "5m"
	Rule1h      = "1h"
	RuleRestart = "until restart"
	RuleAlways  = "always"

	FilterByPath    = "process.path"
	FilterByCommand = "process.command"
	FilterByUserID  = "user.id"
	FilterByDstIP   = "dest.ip"
	FilterByDstPort = "dest.port"
	FilterByDstHost = "dest.host"
)

// NewClient setups a new client and starts the server to listen for new nodes.
func NewClient(serverProto, serverPort string, asDaemon bool, db *storage.Storage) *Client {
	c := &Client{
		db:           db,
		workAsDaemon: asDaemon,
		nodesChan:    make(chan bool),
		rulesInChan:  make(chan *protocol.Connection, 1),
		rulesOutChan: make(chan *protocol.Rule, 1),
	}
	if asDaemon == false {
		go StartServer(c, serverProto, serverPort)
	}

	return c
}

// UpdateStats save latest stats received from a node.
func (c *Client) UpdateStats(ctx context.Context, stats *protocol.Statistics) {
	if stats == nil {
		return
	}
	c.Lock()
	defer c.Unlock()

	nodeAddr := nodes.GetAddr(ctx)
	c.lastStats = stats
	if c.db != nil {
		c.db.Update(nodeAddr, stats)
	}
	nodes.UpdateStats(ctx, stats)
}

// GetStats gets global stats from the db
func (c *Client) GetStats() (stats *[]storage.Statistics) {
	if c.db != nil {
		stats = c.db.GetStats()
	}
	return stats
}

// GetNodeStats gets global stats from the db
func (c *Client) GetNodeStats() (nodes *[]storage.Node) {
	if c.db != nil {
		nodes = c.db.GetNodeStats()
	}
	return nodes
}

// GetEvents gets events from the db
func (c *Client) GetEvents(order string, limit int) (events *[]storage.Connection) {
	if c.db != nil {
		events = c.db.GetEvents(order, limit)
	}
	return events
}

// GetEventsByType returns the list events from the db, by type.
func (c *Client) GetEventsByType(viewType, order string, limit int) (events *[]storage.EventByType) {
	if c.db != nil {
		events = c.db.GetEventsByType(viewType, order, limit)
	}
	return events
}

// GetRules returns the list of rules from the db.
func (c *Client) GetRules(order string, limit int) (rules *[]storage.Rule) {
	if c.db != nil {
		rules = c.db.GetRules(order, limit)
	}
	return rules
}

// GetLastStats returns latest stasts from a node.
func (c *Client) GetLastStats() *protocol.Statistics {
	c.RLock()
	defer c.RUnlock()

	// TODO: return last stats for a given node
	return c.lastStats
}

// AskRule sends the connection details through a channel.
// A client must consume data on that channel, and send the response via the
// rulesOutChan channel.
func (c *Client) AskRule(ctx context.Context, con *protocol.Connection) chan *protocol.Rule {
	if c.workAsDaemon {
		c.rulesOutChan <- nil
		return c.rulesOutChan
	}
	c.rulesInChan <- con
	return c.rulesOutChan
}

// AddNewNode adds a new node to the list of connected nodes.
func (c *Client) AddNewNode(ctx context.Context, nodeConf *protocol.ClientConfig) {
	log.Info("AddNewNode: %s - %s", nodeConf.Name, nodeConf.Version)
	nodes.Add(ctx, nodeConf)
	if c.db != nil {
		c.db.AddNode(nodes.GetAddr(ctx), nodeConf)
	}
	c.nodesChan <- true
}

// OpenChannelWithNode updates the node with the streaming channel.
// This channel is used to send notifications to the nodes (change debug level,
// stop/start interception, etc).
func (c *Client) OpenChannelWithNode(notificationsStream protocol.UI_NotificationsServer) {
	log.Info("opening communication channel with new node...", notificationsStream)
	node := nodes.SetNotificationsChannel(notificationsStream)
	if node == nil {
		log.Warning("node not found, channel comms not opened")
		return
	}
	// XXX: go nodes.Channel(node) ?
	for {
		select {
		case <-node.NotificationsStream.Context().Done():
			log.Important("client.ChannelWithNode() Node exited: ", node.Addr())
			goto Exit
		case notif := <-node.GetNotifications():
			log.Important("client.ChannelWithNode() sending notification:", notif)
			if err := node.NotificationsStream.Send(notif); err != nil {
				log.Error("Error: %v", err)
			}
		}
	}

Exit:
	node.Close()
	return
}

// FIXME: remove when nodes implementation is done
func (c *Client) WaitForNodes() {
	<-c.nodesChan
}

// WaitForRules returns the channel where we listen for new outgoing connections.
func (c *Client) WaitForRules() chan *protocol.Connection {
	return c.rulesInChan
}

// AddNewRule sends a new rule to the node.
func (c *Client) AddNewRule(rule *protocol.Rule) {
	c.rulesOutChan <- rule
}
