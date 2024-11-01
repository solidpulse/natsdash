package natsutil

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/solidpulse/natsdash/ds"
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

func Connect(ctx *ds.NatsCliContext) (*nats.Conn, error) {
	options := []nats.Option{
		nats.Timeout(time.Second * 2),
		nats.MaxReconnects(1),
	}

	if ctx.Token != "" {
		options = append(options, nats.Token(ctx.Token))
	}
	if ctx.User != "" {
		options = append(options, nats.UserInfo(ctx.User, ctx.Password))
	}
	if ctx.Creds != "" {
		options = append(options, nats.UserCredentials(ctx.Creds))
	}
	if ctx.Nkey != "" {
		options = append(options, nats.Nkey(ctx.Nkey, func(nonce []byte) ([]byte, error) {
			// Implement your Nkey signing logic here
			return nil, nil
		}))
	}
	if ctx.Cert != "" && ctx.Key != "" {
		options = append(options, nats.ClientCert(ctx.Cert, ctx.Key))
	}
	if ctx.CA != "" {
		options = append(options, nats.RootCAs(ctx.CA))
	}
	if ctx.NSC != "" {
		options = append(options, nats.Nkey(ctx.NSC, func(nonce []byte) ([]byte, error) {
			// Implement your NSC signing logic here
			return nil, nil
		}))
	}
	if ctx.JetstreamDomain != "" {
		options = append(options, nats.CustomInboxPrefix(ctx.JetstreamDomain))
	}
	if ctx.JetstreamAPIPrefix != "" {
		options = append(options, nats.CustomInboxPrefix(ctx.JetstreamAPIPrefix))
	}
	if ctx.JetstreamEventPrefix != "" {
		options = append(options, nats.CustomInboxPrefix(ctx.JetstreamEventPrefix))
	}
	if ctx.InboxPrefix != "" {
		options = append(options, nats.CustomInboxPrefix(ctx.InboxPrefix))
	}
	// if ctx.UserJWT != "" {
	// 	options = append(options, nats.UserJWT(ctx.UserJWT, func(nonce []byte) ([]byte, error) {
	// 		// Implement your UserJWT signing logic here
	// 		return nil, nil
	// 	}))
	// }

	return nats.Connect(ctx.URL, options...)
}
