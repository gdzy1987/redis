// +build go1.8

package client

import "crypto/tls"

func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	return cfg.Clone()
}
