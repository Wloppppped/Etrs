package internal

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type Option func(cfg *clientv3.Config)

func WithTLSOpt(certFile, keyFile, caFile string) Option {
	return func(cfg *clientv3.Config) {
		tlsCfg, err := newTLSConfig(certFile, keyFile, caFile, "")
		if err != nil {
			zap.L().Error(fmt.Sprintf("tls failed with err: %v , skipping tls.", err))
		}
		cfg.TLS = tlsCfg
	}
}

// WithAuthOpt returns a option that authentication by usernane and password.
func WithAuthOpt(username, password string) Option {
	return func(cfg *clientv3.Config) {
		cfg.Username = username
		cfg.Password = password
	}
}

func newTLSConfig(certFile, keyFile, caFile, serverName string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	successful := caCertPool.AppendCertsFromPEM(caCert)
	if !successful {
		return nil, errors.New("failed to parse ca certificate as PEM encoded content")
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	return cfg, nil
}
