package storage

import (
	"github.com/evilsocket/opensnitch/daemon/ui/protocol"
	"strings"
	"time"
)

func (s *Storage) getProtoStats(addr string, stats *protocol.Statistics) *Statistics {
	return &Statistics{
		Node:          addr,
		DaemonVersion: stats.DaemonVersion,
		Rules:         stats.Rules,
		Uptime:        stats.Uptime,
		DNSResponses:  stats.DnsResponses,
		Connections:   stats.Connections,
		Ignored:       stats.Ignored,
		Accepted:      stats.Accepted,
		Dropped:       stats.Dropped,
		RuleHits:      stats.RuleHits,
		RuleMisses:    stats.RuleMisses,
	}
}

func (s *Storage) getProtoEvents(addr string, stats *protocol.Statistics) (*[]Rule, *[]Connection) {
	var conns []Connection
	var opers []Rule
	for _, ev := range stats.Events {
		opers = append([]Rule{
			Rule{
				Node:     addr,
				Name:     ev.Rule.Name,
				Enabled:  ev.Rule.Enabled,
				Action:   ev.Rule.Action,
				Duration: ev.Rule.Duration,
				Operator: Operator{
					RuleName: ev.Rule.Name,
					Type:     ev.Rule.Operator.Type,
					Operand:  ev.Rule.Operator.Operand,
					Data:     ev.Rule.Operator.Data,
				},
			},
		}, opers...)

		conns = append([]Connection{
			Connection{
				Node:        addr,
				Time:        ev.Time,
				Protocol:    ev.Connection.Protocol,
				SrcIP:       ev.Connection.SrcIp,
				SrcPort:     ev.Connection.SrcPort,
				DstIP:       ev.Connection.DstIp,
				DstHost:     ev.Connection.DstHost,
				DstPort:     ev.Connection.DstPort,
				UserID:      ev.Connection.UserId,
				PID:         ev.Connection.ProcessId,
				ProcessPath: ev.Connection.ProcessPath,
				ProcessCwd:  ev.Connection.ProcessCwd,
				ProcessArgs: strings.Join(ev.Connection.ProcessArgs, " "),
				//ProcessEnv:  ev.Connection.ProcessEnv,
				RuleName:   ev.Rule.Name,
				RuleAction: ev.Rule.Action,
			},
		}, conns...)
		//fmt.Println("getEvents() ", ev)
	}
	return &opers, &conns
}

func (s *Storage) getProtoEventsByType(addr string, stats *protocol.Statistics) *[]EventByType {
	var events []EventByType
	for what, hits := range stats.ByProto {
		events = append(events, []EventByType{
			EventByType{
				Time: time.Now(),
				Node: addr, Name: "ByProto", What: what, Hits: hits},
		}...)
	}
	for what, hits := range stats.ByAddress {
		events = append(events, []EventByType{
			EventByType{
				Time: time.Now(),
				Node: addr, Name: "ByAddress", What: what, Hits: hits},
		}...)
	}
	for what, hits := range stats.ByHost {
		events = append(events, []EventByType{
			EventByType{
				Time: time.Now(),
				Node: addr, Name: "ByHost", What: what, Hits: hits},
		}...)
	}
	for what, hits := range stats.ByPort {
		events = append(events, []EventByType{
			EventByType{
				Time: time.Now(),
				Node: addr, Name: "ByPort", What: what, Hits: hits},
		}...)
	}
	for what, hits := range stats.ByUid {
		events = append(events, []EventByType{
			EventByType{
				Time: time.Now(),
				Node: addr, Name: "ByUid", What: what, Hits: hits},
		}...)
	}
	for what, hits := range stats.ByExecutable {
		events = append(events, []EventByType{
			EventByType{
				Time: time.Now(),
				Node: addr, Name: "ByExecutable", What: what, Hits: hits},
		}...)
	}
	return &events
}

func (s *Storage) getProtoRules(node string, nodeConf *protocol.ClientConfig) *[]Rule {
	var rules []Rule
	for _, pRule := range nodeConf.Rules {
		rules = append(rules,
			[]Rule{
				Rule{
					Node:     node,
					Name:     pRule.Name,
					Enabled:  pRule.Enabled,
					Action:   pRule.Action,
					Duration: pRule.Duration,
					Operator: Operator{
						Type:    pRule.Operator.Type,
						Operand: pRule.Operator.Operand,
						Data:    pRule.Operator.Data,
					},
				},
			}...)
	}

	return &rules
}
