// Package rawendpoint turns CAN frames written to pgn.Writer into RAW format and saves them to a file.
package rawendpoint

import (
	"context"
	"os"

	"github.com/boatkit-io/n2k/pkg/converter"
	"github.com/boatkit-io/n2k/pkg/endpoint"
	"github.com/brutella/can"
	"github.com/sirupsen/logrus"
)

// RawEndpoint writes a raw log file from canbus frames sent through the write pipeline.
// Initially through stdout
type RawEndpoint struct {
	log     *logrus.Logger
	file    *os.File
	handler endpoint.MessageHandler
}

// NewRawEndpoint creates a new RAW endpoint
func NewRawEndpoint(outFilePath string, log *logrus.Logger) *RawEndpoint {
	retval := RawEndpoint{}
	if outFilePath != "" {
		file, err := os.Create(outFilePath)
		if err != nil {
			log.Infof("RAW output file failed to open: %s", err)
		} else {
			retval.file = file
		}

	}
	return &retval
}

// Run method opens the specified log file and kicks off a goroutine that sends frames to the handler
func (r *RawEndpoint) Run(ctx context.Context) error {

	defer r.file.Close()

	go func() {
		<-ctx.Done()
	}()

	return nil
}

// SetOutput sets the output struct for handling when a message is ready
func (r *RawEndpoint) SetOutput(mh endpoint.MessageHandler) {
	r.handler = mh
}

// WriteFrame is invoked by CanAdapter, converts the frame into a RAW string, and writes it to the file.
func (r *RawEndpoint) WriteFrame(frame can.Frame) {
	outStr := converter.RawFromCanFrame(frame)
	if r.file != nil {
		_, _ = r.file.WriteString(outStr)
	} else {
		r.log.Info(outStr)

	}

}
