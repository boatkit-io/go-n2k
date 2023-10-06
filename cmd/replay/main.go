package main

import (
	//	"context"
	"context"
	"flag"
	"os"
	"strings"

	//	"time"

	"github.com/boatkit-io/n2k/pkg/adapter"
	"github.com/boatkit-io/n2k/pkg/adapter/canadapter"
	"github.com/boatkit-io/n2k/pkg/endpoint/n2kendpoint"
	"github.com/boatkit-io/n2k/pkg/pgn"
	"github.com/boatkit-io/n2k/pkg/pkt"
	"github.com/boatkit-io/n2k/pkg/subscribe"

	//	"github.com/boatkit-io/tugboat/pkg/service"
	"github.com/sirupsen/logrus"
)

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// Command-line parsing
	var replayFile string
	flag.StringVar(&replayFile, "replayFile", "", "An optional replay file to run")
	var dumpPgns bool
	flag.BoolVar(&dumpPgns, "dumpPgns", false, "Debug spew all PGNs coming down the pipe")
	flag.Parse()

	log := logrus.StandardLogger()
	log.Infof("in replayfile, dump:%t, file:%s\n", dumpPgns, replayFile)

	subs := subscribe.New()
	go func() {
		if dumpPgns {
			_, _ = subs.SubscribeToAllStructs(func(p interface{}) {
				log.Infof("Handling PGN: %s", pgn.DebugDumpPGN(p))
			})
		}
	}()

	ps := pkt.NewPacketStruct()
	ps.SubscribeToPGNReady(func(fullPGN any) {
		subs.ServeStruct(fullPGN)
	})

	//	ctx, cancel := context.WithCancel(context.Background())
	//	defer cancel()
	if len(replayFile) > 0 && strings.HasSuffix(replayFile, ".n2k") {
		ca := canadapter.NewCanAdapter(log)
		ca.SubscribeToPacketReady(func(p pkt.Packet) {
			ps.ProcessPacket(p)
		})

		ep := n2kendpoint.NewN2kEndpoint(replayFile, log)
		ep.SubscribeToFrameReady(func(m adapter.Message) {
			ca.ProcessMessage(m)
		})

		ctx := context.Background()
		err := ep.Run(ctx)
		if err != nil {
			exitCode = 1
			return
		}
	}
}
