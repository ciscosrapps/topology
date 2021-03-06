package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stephenrlouie/service"
	"github.com/ciscosrapps/topology/config"
	"github.com/ciscosrapps/topology/database"
	"github.com/ciscosrapps/topology/handler"
	"github.com/ciscosrapps/topology/kafka"
	"github.com/ciscosrapps/topology/log"
)

func init() {
	if err := config.InitFlags(TopologyCmd, config.InitTopologyCfg()); err != nil {
		TopologyExit(err)
	}
}

var (
	ErrLocalASNRequired = errors.New("A Valid Local ASN is required")
)

var TopologyCmd = &cobra.Command{
	Use:   "topology",
	Short: "Populate network topology",
	Long:  "Reads kafka BMP messages and populates Arango with the network topology",
	Run:   topologyRun,
}

func TopologyExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "--- Topology encountered an Error: ---\n")
		fmt.Fprintf(os.Stderr, "\t%v\n", err)
		os.Exit(1)
	}
}

func topologyRun(cmd *cobra.Command, args []string) {
	var globalErr error
	defer func() {
		TopologyExit(globalErr)
	}()

	icfg, err := config.GetConfig(config.InitTopologyCfg())
	if err != nil {
		globalErr = err
		return
	}
	cfg := icfg.(*config.TopologyConfig)

	jcfg, err := config.GetConfig(config.InitGlobalCfg())
	if err != nil {
		globalErr = err
		return
	}
	gcfg := jcfg.(*config.GlobalConfig)
	log.NewLogr(gcfg.Log)

	serviceGroup := service.New()
	serviceGroup.HandleSigint(nil)
	var hndlr handler.Handler = handler.NewDefault()
	arangoDB, err := database.NewArango(gcfg.Arango)
	if err != nil {
		globalErr = err
		return
	}
	hndlr = handler.NewArango(arangoDB, cfg.ASN, cfg.DirectPeerASNS, cfg.TransitProviderASNS)

	consumer, err := kafka.New(cfg.Kafka, hndlr)
	if err != nil {
		globalErr = err
		return
	}

	consumer.SetHandler(hndlr)
	serviceGroup.Add(consumer)
	serviceGroup.Start()
	serviceGroup.Wait()
}
