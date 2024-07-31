// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package testpb

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	flagTls = flag.Bool("use_tls", true, "whether all gRPC middleware tests should use tls")

	certPEM []byte
	keyPEM  []byte
)

// InterceptorTestSuite is a testify/Suite that starts a gRPC PingService server and a client.
type InterceptorTestSuite struct {
	suite.Suite

	TestService TestServiceServer
	ServerOpts  []grpc.ServerOption
	ClientOpts  []grpc.DialOption

	serverAddr     string
	ServerListener net.Listener
	Server         *grpc.Server
	clientConn     *grpc.ClientConn
	Client         TestServiceClient

	restartServerWithDelayedStart chan time.Duration
	serverRunning                 chan bool

	cancels []context.CancelFunc
}

func (s *InterceptorTestSuite) SetupSuite() {
	s.restartServerWithDelayedStart = make(chan time.Duration)
	s.serverRunning = make(chan bool)

	s.serverAddr = "127.0.0.1:0"
	var err error
	certPEM, keyPEM, err = generateCertAndKey([]string{"localhost", "example.com"}) // CI:LOCALHOST_OK
	require.NoError(s.T(), err, "unable to generate test certificate/key")

	go func() {
		for {
			var err error
			s.ServerListener, err = net.Listen("tcp", s.serverAddr)
			require.NoError(s.T(), err, "must be able to allocate a port for serverListener")
			s.serverAddr = s.ServerListener.Addr().String()
			if *flagTls {
				cert, err := tls.X509KeyPair(certPEM, keyPEM)
				require.NoError(s.T(), err, "unable to load test TLS certificate")
				creds := credentials.NewServerTLSFromCert(&cert)
				s.ServerOpts = append(s.ServerOpts, grpc.Creds(creds))
			}
			// This is the point where we hook up the interceptor.
			s.Server = grpc.NewServer(s.ServerOpts...)
			// Create a service if the instantiator hasn't provided one.
			if s.TestService == nil {
				s.TestService = &TestPingService{}
			}
			RegisterTestServiceServer(s.Server, s.TestService)

			w := sync.WaitGroup{}
			w.Add(1)
			go func() {
				_ = s.Server.Serve(s.ServerListener)
				w.Done()
			}()
			if s.Client == nil {
				s.Client = s.NewClient(s.ClientOpts...)
			}

			s.serverRunning <- true

			d := <-s.restartServerWithDelayedStart
			s.Server.Stop()
			time.Sleep(d)
			w.Wait()
		}
	}()

	select {
	case <-s.serverRunning:
	case <-time.After(2 * time.Second):
		s.T().Fatal("server failed to start before deadline")
	}
}

func (s *InterceptorTestSuite) RestartServer(delayedStart time.Duration) <-chan bool {
	s.restartServerWithDelayedStart <- delayedStart
	time.Sleep(10 * time.Millisecond)
	return s.serverRunning
}

func (s *InterceptorTestSuite) NewClient(dialOpts ...grpc.DialOption) TestServiceClient {
	//lint:ignore SA1019 This is a vendored package, so we shouldn't be modifying it.
	newDialOpts := append(dialOpts, grpc.WithBlock())
	var err error
	if *flagTls {
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(certPEM) {
			s.T().Fatal("failed to append certificate")
		}
		creds := credentials.NewTLS(&tls.Config{ServerName: "localhost", RootCAs: cp}) // CI:LOCALHOST_OK
		newDialOpts = append(newDialOpts, grpc.WithTransportCredentials(creds))
	} else {
		newDialOpts = append(newDialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//lint:ignore SA1019 This is a vendored package, so we shouldn't be modifying it.
	s.clientConn, err = grpc.DialContext(ctx, s.ServerAddr(), newDialOpts...)
	require.NoError(s.T(), err, "must not error on client Dial")
	return NewTestServiceClient(s.clientConn)
}

func (s *InterceptorTestSuite) ServerAddr() string {
	return s.serverAddr
}

type ctxTestNumber struct{}

var (
	ctxTestNumberKey = &ctxTestNumber{}
	zero             = 0
)

func ExtractCtxTestNumber(ctx context.Context) *int {
	if v, ok := ctx.Value(ctxTestNumberKey).(*int); ok {
		return v
	}
	return &zero
}

// UnaryServerInterceptor returns a new unary server interceptors that adds query information logging.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// newCtx := newContext(ctx, log, opts)
		newCtx := ctx
		resp, err := handler(newCtx, req)
		return resp, err
	}
}

func (s *InterceptorTestSuite) SimpleCtx() context.Context {
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	ctx = context.WithValue(ctx, ctxTestNumberKey, 1)
	s.cancels = append(s.cancels, cancel)
	return ctx
}

func (s *InterceptorTestSuite) DeadlineCtx(deadline time.Time) context.Context {
	ctx, cancel := context.WithDeadline(context.TODO(), deadline)
	s.cancels = append(s.cancels, cancel)
	return ctx
}

func (s *InterceptorTestSuite) TearDownSuite() {
	time.Sleep(10 * time.Millisecond)
	if s.ServerListener != nil {
		s.Server.GracefulStop()
		s.T().Logf("stopped grpc.Server at: %v", s.ServerAddr())
		_ = s.ServerListener.Close()
	}
	if s.clientConn != nil {
		_ = s.clientConn.Close()
	}
	for _, c := range s.cancels {
		c()
	}
}

// generateCertAndKey copied from https://github.com/johanbrandhorst/certify/blob/master/issuers/vault/vault_suite_test.go#L255
// with minor modifications.
func generateCertAndKey(san []string) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "example.com",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              san,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return nil, nil, err
	}
	certOut := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	keyOut := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return certOut, keyOut, nil
}
