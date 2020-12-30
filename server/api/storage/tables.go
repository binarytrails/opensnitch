package storage

import (
	"time"
)

const (
	FieldTime    = "time"
	FieldName    = "name"
	FieldAddress = "address"
	FieldNode    = "node"
	FieldWhat    = "what"
	FieldHits    = "hits"
)

// Node holds the definition of a node.
// A node has its own configuration, version, name (address) and log level
type Node struct {
	//idx               int64  `sql:"AUTO_INCREMENT"`
	Address           string `gorm:"primaryKey;not null"`
	Name              string
	Version           string
	IsFirewallRunning bool
	Config            string
	LogLevel          uint32
	// Node is the field of Connection, and the Node name is saved here under Address
	Connection []Connection `gorm:"foreignKey:Node;references:Address"`
	Statistics Statistics   `gorm:"foreignKey:Node;references:Address"`
}

// Statistics holds the global statistics of the nodes.
// There's no reason to save these stats per minute, so
// if there's already an entry of a given node, we just update the row.
type Statistics struct {
	//idx           int64  `sql:"AUTO_INCREMENT"`
	Node          string `gorm:"index;primaryKey;not null"`
	DaemonVersion string
	Rules         uint64
	Uptime        uint64
	DNSResponses  uint64
	Connections   uint64
	Ignored       uint64
	Accepted      uint64
	Dropped       uint64
	RuleHits      uint64
	RuleMisses    uint64
}

// EventByType holds the information about statistics by type
type EventByType struct {
	Time time.Time
	Name string
	What string `gorm:"primaryKey"`
	Hits uint64
	Node string `gorm:"primaryKey"`
}

// Connection holds the information of a connection event.
// We want to keep the rule name, even if the Rule is deleted from the db.
type Connection struct {
	//ID          int `gorm:"primaryKey"`
	// Unixnano uint64
	Time        string
	Node        string `gorm:"primaryKey"`
	Protocol    string `gorm:"primaryKey"`
	SrcIP       string `gorm:"primaryKey"`
	SrcPort     uint32 `gorm:"primaryKey"`
	DstIP       string `gorm:"primaryKey"`
	DstHost     string `gorm:"primaryKey"`
	DstPort     uint32 `gorm:"primaryKey"`
	UserID      uint32 `gorm:"primaryKey"`
	PID         uint32 `gorm:"primaryKey"`
	ProcessPath string `gorm:"primaryKey"`
	ProcessCwd  string
	ProcessArgs string
	//ProcessEnv  []ProcessEnv `gorm:"type:text"`
	RuleName   string `gorm:"foreignKey"`
	RuleAction string
}

/*type ProcessEnv struct {
	ID    int
	Key   string
	Value string
}*/

// Rule holds the definition of a rule to be applied on a connection.
// The name uniqueness is handled by the daemon, so when we receive the
// rules they should be already unique, thus we don't want duplicated rules.
type Rule struct {
	Name     string `gorm:"primaryKey"`
	Node     string
	Enabled  bool
	Action   string
	Duration string
	Operator Operator `gorm:"foreignKey:RuleName;references:Name;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

// Operator defines the conditions of a rule.
type Operator struct {
	// we don't want duplicated Operators.
	Type     string
	Operand  string
	Data     string
	RuleName string `gorm:"primaryKey"`
}
