package natsutil

import (
	"time"

	"github.com/nats-io/nats.go"
)

func TestConnect(url string) error {
	conn, err := nats.Connect(url,
		nats.Timeout(time.Second*2),
		nats.MaxReconnects(1))
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func Connect(url string) (*nats.Conn, error) {
	return nats.Connect(url)
}
