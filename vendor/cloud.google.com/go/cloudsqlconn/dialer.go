// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudsqlconn

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"cloud.google.com/go/cloudsqlconn/errtype"
	"cloud.google.com/go/cloudsqlconn/internal/cloudsql"
	"cloud.google.com/go/cloudsqlconn/internal/trace"
	"github.com/google/uuid"
	"golang.org/x/net/proxy"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

const (
	// defaultTCPKeepAlive is the default keep alive value used on connections to a Cloud SQL instance.
	defaultTCPKeepAlive = 30 * time.Second
	// serverProxyPort is the port the server-side proxy receives connections on.
	serverProxyPort = "3307"
	// iamLoginScope is the OAuth2 scope used for tokens embedded in the ephemeral
	// certificate.
	iamLoginScope = "https://www.googleapis.com/auth/sqlservice.login"
)

var (
	// versionString indicates the version of this library.
	//go:embed version.txt
	versionString string
	userAgent     = "cloud-sql-go-connector/" + strings.TrimSpace(versionString)

	// defaultKey is the default RSA public/private keypair used by the clients.
	defaultKey    *rsa.PrivateKey
	defaultKeyErr error
	keyOnce       sync.Once
)

func getDefaultKeys() (*rsa.PrivateKey, error) {
	keyOnce.Do(func() {
		defaultKey, defaultKeyErr = rsa.GenerateKey(rand.Reader, 2048)
	})
	return defaultKey, defaultKeyErr
}

// A Dialer is used to create connections to Cloud SQL instances.
//
// Use NewDialer to initialize a Dialer.
type Dialer struct {
	lock sync.RWMutex
	// instances map connection names (e.g., my-project:us-central1:my-instance)
	// to *cloudsql.Instance types.
	instances      map[cloudsql.ConnName]*cloudsql.Instance
	key            *rsa.PrivateKey
	refreshTimeout time.Duration

	sqladmin *sqladmin.Service

	// defaultDialCfg holds the constructor level DialOptions, so that it can
	// be copied and mutated by the Dial function.
	defaultDialCfg dialCfg

	// dialerID uniquely identifies a Dialer. Used for monitoring purposes,
	// *only* when a client has configured OpenCensus exporters.
	dialerID string

	// dialFunc is the function used to connect to the address on the named
	// network. By default, it is golang.org/x/net/proxy#Dial.
	dialFunc func(cxt context.Context, network, addr string) (net.Conn, error)

	// iamTokenSource supplies the OAuth2 token used for IAM DB Authn.
	iamTokenSource oauth2.TokenSource
}

var (
	errUseTokenSource    = errors.New("use WithTokenSource when IAM AuthN is not enabled")
	errUseIAMTokenSource = errors.New("use WithIAMAuthNTokenSources instead of WithTokenSource be used when IAM AuthN is enabled")
)

// NewDialer creates a new Dialer.
//
// Initial calls to NewDialer make take longer than normal because generation of an
// RSA keypair is performed. Calls with a WithRSAKeyPair DialOption or after a default
// RSA keypair is generated will be faster.
func NewDialer(ctx context.Context, opts ...Option) (*Dialer, error) {
	cfg := &dialerConfig{
		refreshTimeout: cloudsql.RefreshTimeout,
		dialFunc:       proxy.Dial,
		useragents:     []string{userAgent},
	}
	for _, opt := range opts {
		opt(cfg)
		if cfg.err != nil {
			return nil, cfg.err
		}
	}
	if cfg.useIAMAuthN && cfg.setTokenSource && !cfg.setIAMAuthNTokenSource {
		return nil, errUseIAMTokenSource
	}
	if cfg.setIAMAuthNTokenSource && !cfg.useIAMAuthN {
		return nil, errUseTokenSource
	}
	// Add this to the end to make sure it's not overridden
	cfg.sqladminOpts = append(cfg.sqladminOpts, option.WithUserAgent(strings.Join(cfg.useragents, " ")))

	// If callers have not provided a token source, either explicitly with
	// WithTokenSource or implicitly with WithCredentialsJSON etc., then use the
	// default token source.
	if !cfg.setCredentials {
		ts, err := google.DefaultTokenSource(ctx, sqladmin.SqlserviceAdminScope)
		if err != nil {
			return nil, fmt.Errorf("failed to create token source: %v", err)
		}
		cfg.sqladminOpts = append(cfg.sqladminOpts, option.WithTokenSource(ts))
		scoped, err := google.DefaultTokenSource(ctx, iamLoginScope)
		if err != nil {
			return nil, fmt.Errorf("failed to create scoped token source: %v", err)
		}
		cfg.iamLoginTokenSource = scoped
	}

	if cfg.rsaKey == nil {
		key, err := getDefaultKeys()
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA keys: %v", err)
		}
		cfg.rsaKey = key
	}

	client, err := sqladmin.NewService(ctx, cfg.sqladminOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqladmin client: %v", err)
	}

	dc := dialCfg{
		ipType:       cloudsql.PublicIP,
		tcpKeepAlive: defaultTCPKeepAlive,
		refreshCfg: cloudsql.RefreshCfg{
			UseIAMAuthN: cfg.useIAMAuthN,
		},
	}
	for _, opt := range cfg.dialOpts {
		opt(&dc)
	}

	if err := trace.InitMetrics(); err != nil {
		return nil, err
	}
	d := &Dialer{
		instances:      make(map[cloudsql.ConnName]*cloudsql.Instance),
		key:            cfg.rsaKey,
		refreshTimeout: cfg.refreshTimeout,
		sqladmin:       client,
		defaultDialCfg: dc,
		dialerID:       uuid.New().String(),
		iamTokenSource: cfg.iamLoginTokenSource,
		dialFunc:       cfg.dialFunc,
	}
	return d, nil
}

// Dial returns a net.Conn connected to the specified Cloud SQL instance. The instance argument must be the
// instance's connection name, which is in the format "project-name:region:instance-name".
func (d *Dialer) Dial(ctx context.Context, instance string, opts ...DialOption) (conn net.Conn, err error) {
	startTime := time.Now()
	var endDial trace.EndSpanFunc
	ctx, endDial = trace.StartSpan(ctx, "cloud.google.com/go/cloudsqlconn.Dial",
		trace.AddInstanceName(instance),
		trace.AddDialerID(d.dialerID),
	)
	defer func() {
		go trace.RecordDialError(context.Background(), instance, d.dialerID, err)
		endDial(err)
	}()
	cn, err := cloudsql.ParseConnName(instance)
	if err != nil {
		return nil, err
	}

	cfg := d.defaultDialCfg
	for _, opt := range opts {
		opt(&cfg)
	}

	var endInfo trace.EndSpanFunc
	ctx, endInfo = trace.StartSpan(ctx, "cloud.google.com/go/cloudsqlconn/internal.InstanceInfo")
	i := d.instance(cn, &cfg.refreshCfg)
	addr, tlsCfg, err := i.ConnectInfo(ctx, cfg.ipType)
	if err != nil {
		d.removeInstance(i)
		endInfo(err)
		return nil, err
	}
	endInfo(err)

	var connectEnd trace.EndSpanFunc
	ctx, connectEnd = trace.StartSpan(ctx, "cloud.google.com/go/cloudsqlconn/internal.Connect")
	defer func() { connectEnd(err) }()
	addr = net.JoinHostPort(addr, serverProxyPort)
	f := d.dialFunc
	if cfg.dialFunc != nil {
		f = cfg.dialFunc
	}
	conn, err = f(ctx, "tcp", addr)
	if err != nil {
		// refresh the instance info in case it caused the connection failure
		i.ForceRefresh()
		return nil, errtype.NewDialError("failed to dial", i.String(), err)
	}
	if c, ok := conn.(*net.TCPConn); ok {
		if err := c.SetKeepAlive(true); err != nil {
			return nil, errtype.NewDialError("failed to set keep-alive", i.String(), err)
		}
		if err := c.SetKeepAlivePeriod(cfg.tcpKeepAlive); err != nil {
			return nil, errtype.NewDialError("failed to set keep-alive period", i.String(), err)
		}
	}

	tlsConn := tls.Client(conn, tlsCfg)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		// refresh the instance info in case it caused the handshake failure
		i.ForceRefresh()
		_ = tlsConn.Close() // best effort close attempt
		return nil, errtype.NewDialError("handshake failed", i.String(), err)
	}

	latency := time.Since(startTime).Milliseconds()
	go func() {
		n := atomic.AddUint64(&i.OpenConns, 1)
		trace.RecordOpenConnections(ctx, int64(n), d.dialerID, i.String())
		trace.RecordDialLatency(ctx, instance, d.dialerID, latency)
	}()

	return newInstrumentedConn(tlsConn, func() {
		n := atomic.AddUint64(&i.OpenConns, ^uint64(0))
		trace.RecordOpenConnections(context.Background(), int64(n), d.dialerID, i.String())
	}), nil
}

// EngineVersion returns the engine type and version for the instance. The value will
// correspond to one of the following types for the instance:
// https://cloud.google.com/sql/docs/mysql/admin-api/rest/v1beta4/SqlDatabaseVersion
func (d *Dialer) EngineVersion(ctx context.Context, instance string) (string, error) {
	cn, err := cloudsql.ParseConnName(instance)
	if err != nil {
		return "", err
	}
	i := d.instance(cn, nil)
	e, err := i.InstanceEngineVersion(ctx)
	if err != nil {
		return "", err
	}
	return e, nil
}

// Warmup starts the background refresh necessary to connect to the instance. Use Warmup
// to start the refresh process early if you don't know when you'll need to call "Dial".
func (d *Dialer) Warmup(_ context.Context, instance string, opts ...DialOption) error {
	cn, err := cloudsql.ParseConnName(instance)
	if err != nil {
		return err
	}
	cfg := d.defaultDialCfg
	for _, opt := range opts {
		opt(&cfg)
	}
	_ = d.instance(cn, &cfg.refreshCfg)
	return nil
}

// newInstrumentedConn initializes an instrumentedConn that on closing will
// decrement the number of open connects and record the result.
func newInstrumentedConn(conn net.Conn, closeFunc func()) *instrumentedConn {
	return &instrumentedConn{
		Conn:      conn,
		closeFunc: closeFunc,
	}
}

// instrumentedConn wraps a net.Conn and invokes closeFunc when the connection
// is closed.
type instrumentedConn struct {
	net.Conn
	closeFunc func()
}

// Close delegates to the underlying net.Conn interface and reports the close
// to the provided closeFunc only when Close returns no error.
func (i *instrumentedConn) Close() error {
	err := i.Conn.Close()
	if err != nil {
		return err
	}
	go i.closeFunc()
	return nil
}

// Close closes the Dialer; it prevents the Dialer from refreshing the information
// needed to connect. Additional dial operations may succeed until the information
// expires.
func (d *Dialer) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, i := range d.instances {
		i.Close()
	}
	return nil
}

// instance is a helper function for returning the appropriate instance object in a threadsafe way.
// It will create a new instance object, modify the existing one, or leave it unchanged as needed.
func (d *Dialer) instance(cn cloudsql.ConnName, r *cloudsql.RefreshCfg) *cloudsql.Instance {
	// Check instance cache
	d.lock.RLock()
	i, ok := d.instances[cn]
	d.lock.RUnlock()
	// If the instance hasn't been created yet or if the refreshCfg has changed
	if !ok || (r != nil && *r != i.RefreshCfg) {
		d.lock.Lock()
		// Recheck to ensure instance wasn't created or changed between locks
		i, ok = d.instances[cn]
		if !ok {
			// Create a new instance
			if r == nil {
				r = &d.defaultDialCfg.refreshCfg
			}
			i = cloudsql.NewInstance(cn, d.sqladmin, d.key,
				d.refreshTimeout, d.iamTokenSource, d.dialerID, *r)
			d.instances[cn] = i
		} else if r != nil && *r != i.RefreshCfg {
			// Update the instance with the new refresh cfg
			i.UpdateRefresh(*r)
		}
		d.lock.Unlock()
	}
	return i
}

func (d *Dialer) removeInstance(i *cloudsql.Instance) {
	d.lock.Lock()
	defer d.lock.Unlock()
	// Stop all background refreshes
	i.Close()
	delete(d.instances, i.ConnName)
}
