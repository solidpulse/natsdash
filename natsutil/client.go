package natsutil

import (
	"time"

	"github.com/nats-io/nats.go"
)

func TestConnect() error {
	conn, err := nats.Connect("nats://localhWost:4222",
		nats.Timeout(time.Second*2),
		nats.MaxReconnects(1))
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
