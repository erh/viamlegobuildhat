package main

import (
	"github.com/erh/viambuildhat"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {

	conn, err := viambuildhat.NewConnection("/dev/serial0")
	if err != nil {
		return err
	}

	defer conn.Close()

	return nil

}
