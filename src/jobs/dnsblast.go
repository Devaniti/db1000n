package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Arriven/db1000n/src/dnsblast"
	"github.com/Arriven/db1000n/src/utils"
)

const (
	defaultProto                   = dnsblast.UDPProtoName
	defaultParallelQueriesPerCycle = 5
	defaultIntervalInMS            = 10
)

type dnsBlastConfig struct {
	BasicJobConfig
	RootDomain      string   `json:"root_domain"`
	Protocol        string   `json:"protocol"` // "udp", "tcp", "tcp-tls"
	SeedDomains     []string `json:"seed_domains"`
	ParallelQueries int      `json:"parallel_queries"`
}

func dnsBlastJob(ctx context.Context, args Args, debug bool) error {
	defer utils.PanicHandler()

	jobConfig := new(dnsBlastConfig)
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return fmt.Errorf("failed to parse DNS Blast job configurations: %s", err)
	}

	//
	// Default settings and early misconfiguration prevention
	//

	// Root domain verification
	if len(jobConfig.RootDomain) == 0 {
		return errors.New("no root domain provided, consider adding it")
	}

	// Domain seeds verification
	if len(jobConfig.SeedDomains) == 0 {
		return errors.New("no seed domains provided, at least one is required")
	}

	// Protocol settlement
	isUDPProto := jobConfig.Protocol == dnsblast.UDPProtoName
	isTCPProto := jobConfig.Protocol == dnsblast.TCPProtoName
	isTCPTLSProto := jobConfig.Protocol == dnsblast.TCPTLSProtoName

	switch {

	case jobConfig.Protocol == "":
		jobConfig.Protocol = defaultProto
		isUDPProto = true

	case !(isUDPProto || !isTCPProto || !isTCPTLSProto):
		return fmt.Errorf("unrecognized DNS protocol [provided], expected one of [%v]",
			[]string{dnsblast.UDPProtoName, dnsblast.TCPProtoName, dnsblast.TCPTLSProtoName})
	}

	// Concurrency validation
	if jobConfig.ParallelQueries == 0 {
		jobConfig.ParallelQueries = defaultParallelQueriesPerCycle
	}

	// Delay validation
	if jobConfig.IntervalMs == 0 {
		jobConfig.IntervalMs = defaultIntervalInMS
	}

	//
	// Blast the Job!
	//
	return dnsblast.Start(ctx, &dnsblast.Config{
		RootDomain:      jobConfig.RootDomain,
		Protocol:        jobConfig.Protocol,
		SeedDomains:     jobConfig.SeedDomains,
		ParallelQueries: jobConfig.ParallelQueries,
		Delay:           time.Duration(jobConfig.IntervalMs) * time.Millisecond,
	})
}
