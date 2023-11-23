package main

import (
	"context"
	"time"

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
	ctx := context.Background()
	logger := logging.NewDebugLogger("motortype")

	conn, err := viambuildhat.NewConnection("/dev/serial0", logger)
	if err != nil {
		return err
	}
	defer conn.Close()

	m, err := viambuildhat.NewMotor(conn, 3, logger)
	if err != nil {
		return err
	}
	defer m.Close(ctx)

	err = m.SetPower(ctx, 1, nil)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 2)

	err = m.Stop(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
