package auth

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/weedge/lib/container/slice"
	"github.com/weedge/lib/log"

	"github.com/Shopify/sarama"
)

type AuthOptions struct {
	// security-protocol: SSL (broker service config default PLAINTEXT, need security.inter.broker.protocol=SSL)
	enableSSL bool
	certFile  string // the optional certificate file for client authentication
	keyFile   string // the optional key file for client authentication
	caFile    string // the optional certificate authority file for TLS client authentication
	verifySSL bool   // the optional verify ssl certificates chain

	//SASL: PLAINTEXT(Kafka 0.10.0.0+), SCRAM(kafka 0.10.2.0+ dynamic add user), OAUTHBEARER(Kafka2.0.0+,JWT)
	enableSASL     bool
	authentication string // PLAINTEXT,SCRAM,OAUTHBEARER,if enableSASL true,default SCRAM
	saslUser       string // The SASL username
	saslPassword   string // The SASL password
	scramAlgorithm string // The SASL SCRAM SHA algorithm sha256 or sha512 as mechanism
}

type Option interface {
	apply(o *AuthOptions)
}

type funcServerOption struct {
	f func(o *AuthOptions)
}

func (fdo *funcServerOption) apply(o *AuthOptions) {
	fdo.f(o)
}

func newFuncServerOption(f func(o *AuthOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

func WithEnableSSL(enableSSL bool) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.enableSSL = enableSSL
	})
}

func WithSSLCertFile(certFile string) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.certFile = certFile
	})
}

func WithSSLKeyFile(keyFile string) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.keyFile = keyFile
	})
}

func WithSSLCaFile(caFile string) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.caFile = caFile
	})
}

func (opts *AuthOptions) InitSSL(config *sarama.Config) {
	if opts.enableSSL == false {
		return
	}

	tlsConfig := opts.CreateTlsConfiguration()
	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}
}

func (opts *AuthOptions) CreateTlsConfiguration() (t *tls.Config) {
	if opts.certFile == "" || opts.keyFile == "" || opts.caFile == "" {
		return
	}

	cert, err := tls.LoadX509KeyPair(opts.certFile, opts.keyFile)
	if err != nil {
		log.Error("tls.LoadX509KeyPair err", err)
		return
	}

	caCert, err := ioutil.ReadFile(opts.caFile)
	if err != nil {
		log.Error("ioutil.ReadFile err", err)
		return
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	t = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: opts.verifySSL,
	}

	return
}

func WithEnableSASL(enableSASL bool) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.enableSASL = enableSASL
	})
}

func WithAuthentication(authentication string) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.authentication = "SCRAM"
		if slice.Contains([]string{"PLAINTEXT", "SCRAM", "OAUTHBEARER"}, authentication) {
			o.authentication = authentication
		}
	})
}

func WithScramAlgorithm(scramAlgorithm string) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.scramAlgorithm = "sha256"
		if slice.Contains([]string{"sha256", "sha512"}, scramAlgorithm) {
			o.scramAlgorithm = scramAlgorithm
		}
	})
}

func WithSASLUser(saslUser string) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.saslUser = saslUser
	})
}

func WithSASLPassword(saslPassword string) Option {
	return newFuncServerOption(func(o *AuthOptions) {
		o.saslPassword = saslPassword
	})
}

func (opts *AuthOptions) InitSASLSCRAM(conf *sarama.Config) {
	if !opts.enableSASL {
		return
	}
	if opts.authentication != "SCRAM" {
		return
	}
	conf.Net.SASL.Enable = opts.enableSASL
	conf.Net.SASL.User = opts.saslUser
	conf.Net.SASL.Password = opts.saslPassword
	conf.Net.SASL.Handshake = true
	switch opts.scramAlgorithm {
	case "sha512":
		conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
		conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	case "sha256":
		fallthrough
	default:
		conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA256} }
		conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	}
}

func GetAuthOptions(opts ...Option) *AuthOptions {
	authOptions := &AuthOptions{
		enableSSL:      false,
		certFile:       "",
		keyFile:        "",
		caFile:         "",
		verifySSL:      false,

		enableSASL:     false,
		authentication: "SCRAM",
		saslUser:       "",
		saslPassword:   "",
		scramAlgorithm: "sha256",
	}

	for _, o := range opts {
		o.apply(authOptions)
	}

	return authOptions
}
