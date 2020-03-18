package suricatalogs

/**
 * Panther is a scalable, powerful, cloud-native SIEM written in Golang/React.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers"
	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers/timestamp"
)

var DNSDesc = `Suricata parser for the DNS event type in the EVE JSON output.
Reference: https://suricata.readthedocs.io/en/suricata-5.0.2/output/eve/eve-json-output.html`

//nolint:lll
type DNS struct {
	CommunityID  *string     `json:"community_id,omitempty" description:"Suricata DNS CommunityID"`
	DNS          *DNSDetails `json:"dns" validate:"required,dive" description:"Suricata DNS DNS"`
	DestIP       *string     `json:"dest_ip" validate:"required" description:"Suricata DNS DestIP"`
	DestPort     *int        `json:"dest_port,omitempty" description:"Suricata DNS DestPort"`
	EventType    *string     `json:"event_type" validate:"required" description:"Suricata DNS EventType"`
	FlowID       *int        `json:"flow_id,omitempty" description:"Suricata DNS FlowID"`
	PcapCnt      *int        `json:"pcap_cnt,omitempty" description:"Suricata DNS PcapCnt"`
	PcapFilename *string     `json:"pcap_filename,omitempty" description:"Suricata DNS PcapFilename"`
	Proto        *string     `json:"proto" validate:"required" description:"Suricata DNS Proto"`
	SrcIP        *string     `json:"src_ip" validate:"required" description:"Suricata DNS SrcIP"`
	SrcPort      *int        `json:"src_port,omitempty" description:"Suricata DNS SrcPort"`
	Timestamp    *string     `json:"timestamp" validate:"required" description:"Suricata DNS Timestamp"`
	Vlan         []int       `json:"vlan,omitempty" description:"Suricata DNS Vlan"`

	parsers.PantherLog
}

//nolint:lll
type DNSDetails struct {
	Aa          *bool                   `json:"aa,omitempty" description:"Suricata DNSDetails Aa"`
	Answers     []DNSDetailsAnswers     `json:"answers,omitempty" validate:"omitempty,dive" description:"Suricata DNSDetails Answers"`
	Authorities []DNSDetailsAuthorities `json:"authorities,omitempty" validate:"omitempty,dive" description:"Suricata DNSDetails Authorities"`
	Flags       *string                 `json:"flags,omitempty" description:"Suricata DNSDetails Flags"`
	Grouped     *DNSDetailsGrouped      `json:"grouped,omitempty" validate:"omitempty,dive" description:"Suricata DNSDetails Grouped"`
	ID          *int                    `json:"id,omitempty" description:"Suricata DNSDetails ID"`
	Qr          *bool                   `json:"qr,omitempty" description:"Suricata DNSDetails Qr"`
	Ra          *bool                   `json:"ra,omitempty" description:"Suricata DNSDetails Ra"`
	Rcode       *string                 `json:"rcode,omitempty" description:"Suricata DNSDetails Rcode"`
	Rd          *bool                   `json:"rd,omitempty" description:"Suricata DNSDetails Rd"`
	Rrname      *string                 `json:"rrname,omitempty" description:"Suricata DNSDetails Rrname"`
	Rrtype      *string                 `json:"rrtype,omitempty" description:"Suricata DNSDetails Rrtype"`
	TxID        *int                    `json:"tx_id,omitempty" description:"Suricata DNSDetails TxID"`
	Type        *string                 `json:"type,omitempty" description:"Suricata DNSDetails Type"`
	Version     *int                    `json:"version,omitempty" description:"Suricata DNSDetails Version"`
}

//nolint:lll
type DNSDetailsAnswers struct {
	Rdata  *string `json:"rdata,omitempty" description:"Suricata DNSDetailsAnswers Rdata"`
	Rrname *string `json:"rrname,omitempty" description:"Suricata DNSDetailsAnswers Rrname"`
	Rrtype *string `json:"rrtype,omitempty" description:"Suricata DNSDetailsAnswers Rrtype"`
	TTL    *int    `json:"ttl,omitempty" description:"Suricata DNSDetailsAnswers TTL"`
}

//nolint:lll
type DNSDetailsGrouped struct {
	A     []string `json:"A,omitempty" description:"Suricata DNSDetailsGrouped A"`
	Aaaa  []string `json:"AAAA,omitempty" description:"Suricata DNSDetailsGrouped Aaaa"`
	Cname []string `json:"CNAME,omitempty" description:"Suricata DNSDetailsGrouped Cname"`
	Mx    []string `json:"MX,omitempty" description:"Suricata DNSDetailsGrouped Mx"`
	Ptr   []string `json:"PTR,omitempty" description:"Suricata DNSDetailsGrouped Ptr"`
	Txt   []string `json:"TXT,omitempty" description:"Suricata DNSDetailsGrouped Txt"`
}

//nolint:lll
type DNSDetailsAuthorities struct {
	Rrname *string `json:"rrname,omitempty" description:"Suricata DNSDetailsAuthorities Rrname"`
	Rrtype *string `json:"rrtype,omitempty" description:"Suricata DNSDetailsAuthorities Rrtype"`
	TTL    *int    `json:"ttl,omitempty" description:"Suricata DNSDetailsAuthorities TTL"`
}

// DNSParser parses Suricata DNS alerts in the JSON format
type DNSParser struct{}

func (p *DNSParser) New() parsers.LogParser {
	return &DNSParser{}
}

// Parse returns the parsed events or nil if parsing failed
func (p *DNSParser) Parse(log string) []*parsers.PantherLog {
	event := &DNS{}

	err := jsoniter.UnmarshalFromString(log, event)
	if err != nil {
		zap.L().Debug("failed to parse log", zap.Error(err))
		return nil
	}

	event.updatePantherFields(p)

	if err := parsers.Validator.Struct(event); err != nil {
		zap.L().Debug("failed to validate log", zap.Error(err))
		return nil
	}

	return event.Logs()
}

// LogType returns the log type supported by this parser
func (p *DNSParser) LogType() string {
	return "Suricata.DNS"
}

func (event *DNS) updatePantherFields(p *DNSParser) {
	eventTime, _ := timestamp.Parse(time.RFC3339Nano, *event.Timestamp)
	event.SetCoreFields(p.LogType(), &eventTime, event)
	event.AppendAnyIPAddressPtrs(event.SrcIP, event.DestIP)
}
