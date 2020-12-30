package storage

import (
	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type DbType int

// Db supported types
const (
	Sqlite = iota
	Postgres
	MySQL
)

// Storage holds the information of the db configured (type and name/dsn)
type Storage struct {
	db       *gorm.DB
	Filename string
}

// NewStorage opens a new connection to the db configured, and creates
// or migrates the tables
func NewStorage(filename string, dbType DbType, dbDebug bool) *Storage {
	var dbT gorm.Dialector

	logLevel := logger.Silent
	if dbDebug {
		logLevel = logger.Info
	}

	switch dbType {
	case Sqlite:
		dbT = sqlite.Open(filename)
	case Postgres:
		dbT = postgres.Open(filename)
	case MySQL:
		dbT = mysql.Open(filename)
	}

	db, err := gorm.Open(dbT, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil
	}

	if dbType == Sqlite {
		db.Exec("PRAGMA foreign_keys = ON;")
	}

	// create the tables
	db.AutoMigrate(&Node{})
	db.AutoMigrate(&Connection{})
	db.AutoMigrate(&Rule{})
	db.AutoMigrate(&Operator{})
	db.AutoMigrate(&Statistics{})
	db.AutoMigrate(&EventByType{})
	return &Storage{
		db:       db,
		Filename: filename,
	}
}

// AddNode adds a new node to the database
func (s *Storage) AddNode(addr string, nodeConfig *protocol.ClientConfig) {
	node := &Node{
		Address:           addr,
		Name:              nodeConfig.Name,
		Version:           nodeConfig.Version,
		IsFirewallRunning: nodeConfig.IsFirewallRunning,
		Config:            nodeConfig.Config,
		LogLevel:          nodeConfig.LogLevel,
	}
	s.db.Model(&Node{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: FieldAddress},
		},
		UpdateAll: true,
	}).Create(node)
	rules := s.getProtoRules(addr, nodeConfig)
	s.AddRules(addr, rules)
}

// AddRules adds rules to the db
func (s *Storage) AddRules(node string, rules *[]Rule) {
	s.db.Model(&Rule{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: FieldName},
		},
		DoNothing: true,
	}).Create(rules)
}

// Update updates the statistics of a node
func (s *Storage) Update(addr string, stats *protocol.Statistics) {
	s.db.Model(&Statistics{}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(s.getProtoStats(addr, stats))
	//s.db.Model(&Event{}).Create(s.getEvents(stats))

	rules, conns := s.getProtoEvents(addr, stats)
	s.AddRules(addr, rules)

	s.db.Model(&Connection{}).Clauses(clause.OnConflict{
		// The order of these fields must match the order of the table fields
		Columns: []clause.Column{
			{Name: FieldNode},
			{Name: "protocol"},
			{Name: "src_ip"},
			{Name: "src_port"},
			{Name: "dst_ip"},
			{Name: "dst_host"},
			{Name: "dst_port"},
			{Name: "user_id"},
			{Name: "p_id"},
			{Name: "process_path"},
		},
		//DoUpdates: clause.AssignmentColumns([]string{"time"}),
		DoNothing: true,
	}).Create(conns)

	s.db.Model(&EventByType{}).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: FieldNode},
			{Name: FieldWhat},
		},
		DoUpdates: clause.AssignmentColumns([]string{FieldHits}),
	}).Create(s.getProtoEventsByType(addr, stats))
}

// GetStats gets the connection statistics ordered and limited if configured.
func (s *Storage) GetStats() *[]Statistics {
	var stats []Statistics
	s.db.Find(&stats)

	return &stats
}

// GetNodeStats gets the connection statistics ordered and limited if configured.
func (s *Storage) GetNodeStats() *[]Node {
	var nodes []Node
	s.db.Joins("Statistics").Find(&nodes)

	return &nodes
}

// GetEvents gets the connection events ordered and limited if configured.
func (s *Storage) GetEvents(order string, limit int) *[]Connection {
	var conns []Connection
	s.db.Order(FieldTime + " " + order).Limit(limit).Find(&conns)

	return &conns
}

// GetEventsByType gets the statistics by type (ports, procs, protocol, user id, address or hosts)
func (s *Storage) GetEventsByType(viewType, order string, limit int) *[]EventByType {
	var events []EventByType
	s.db.Where(FieldName+" = ?", viewType).Order(FieldHits + " " + order).Limit(limit).Find(&events)

	return &events
}

// GetRules gets the rules in use by the daemon
func (s *Storage) GetRules(order string, limit int) *[]Rule {
	var rules []Rule
	s.db.Order(FieldName + " " + order).Limit(limit).Find(&rules)

	return &rules
}
