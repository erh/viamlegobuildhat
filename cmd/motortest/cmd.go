package main

import (
	"go.viam.com/rdk/logging"

	"github.com/erh/viambuildhat"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	logger := logging.NewDebugLogger("motortype")

	conn, err := viambuildhat.NewConnection("/dev/serial0", logger)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}
