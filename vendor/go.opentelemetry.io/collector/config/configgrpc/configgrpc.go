// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configgrpc // import "go.opentelemetry.io/collector/config/configgrpc"

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mostynb/go-grpc-compression/nonclobbering/snappy"
	"github.com/mostynb/go-grpc-compression/nonclobbering/zstd"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/config/internal"
	"go.opentelemetry.io/collector/extension/auth"
)

var errMetadataNotFound = errors.New("no request metadata found")

// KeepaliveClientConfig exposes the keepalive.ClientParameters to be used by the exporter.
// Refer to the original data-structure for the meaning of each parameter:
// https://godoc.org/google.golang.org/grpc/keepalive#ClientParameters
type KeepaliveClientConfig struct {
	Time                time.Duration `mapstructure:"time"`
	Timeout             time.Duration `mapstructure:"timeout"`
	PermitWithoutStream bool          `mapstructure:"permit_without_stream"`
}

// NewDefaultKeepaliveClientConfig returns a new instance of KeepaliveClientConfig with default values.
func NewDefaultKeepaliveClientConfig() *KeepaliveClientConfig {
	return &KeepaliveClientConfig{
		Time:    time.Second * 10,
		Timeout: time.Second * 10,
	}
}

// ClientConfig defines common settings for a gRPC client configuration.
type ClientConfig struct {
	// The target to which the exporter is going to send traces or metrics,
	// using the gRPC protocol. The valid syntax is described at
	// https://github.com/grpc/grpc/blob/master/doc/naming.md.
	Endpoint string `mapstructure:"endpoint"`

	// The compression key for supported compression types within collector.
	Compression configcompression.Type `mapstructure:"compression"`

	// TLSSetting struct exposes TLS client configuration.
	TLSSetting configtls.ClientConfig `mapstructure:"tls"`

	// The keepalive parameters for gRPC client. See grpc.WithKeepaliveParams.
	// (https://godoc.org/google.golang.org/grpc#WithKeepaliveParams).
	Keepalive *KeepaliveClientConfig `mapstructure:"keepalive"`

	// ReadBufferSize for gRPC client. See grpc.WithReadBufferSize.
	// (https://godoc.org/google.golang.org/grpc#WithReadBufferSize).
	ReadBufferSize int `mapstructure:"read_buffer_size"`

	// WriteBufferSize for gRPC gRPC. See grpc.WithWriteBufferSize.
	// (https://godoc.org/google.golang.org/grpc#WithWriteBufferSize).
	WriteBufferSize int `mapstructure:"write_buffer_size"`

	// WaitForReady parameter configures client to wait for ready state before sending data.
	// (https://github.com/grpc/grpc/blob/master/doc/wait-for-ready.md)
	WaitForReady bool `mapstructure:"wait_for_ready"`

	// The headers associated with gRPC requests.
	Headers map[string]configopaque.String `mapstructure:"headers"`

	// Sets the balancer in grpclb_policy to discover the servers. Default is pick_first.
	// https://github.com/grpc/grpc-go/blob/master/examples/features/load_balancing/README.md
	BalancerName string `mapstructure:"balancer_name"`

	// WithAuthority parameter configures client to rewrite ":authority" header
	// (godoc.org/google.golang.org/grpc#WithAuthority)
	Authority string `mapstructure:"authority"`

	// Auth configuration for outgoing RPCs.
	Auth *configauth.Authentication `mapstructure:"auth"`
}

// NewDefaultClientConfig returns a new instance of ClientConfig with default values.
func NewDefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		TLSSetting: configtls.NewDefaultClientConfig(),
		Keepalive:  NewDefaultKeepaliveClientConfig(),
		Auth:       configauth.NewDefaultAuthentication(),
	}
}

// KeepaliveServerConfig is the configuration for keepalive.
type KeepaliveServerConfig struct {
	ServerParameters  *KeepaliveServerParameters  `mapstructure:"server_parameters"`
	EnforcementPolicy *KeepaliveEnforcementPolicy `mapstructure:"enforcement_policy"`
}

// NewDefaultKeepaliveServerConfig returns a new instance of KeepaliveServerConfig with default values.
func NewDefaultKeepaliveServerConfig() *KeepaliveServerConfig {
	return &KeepaliveServerConfig{
		ServerParameters:  NewDefaultKeepaliveServerParameters(),
		EnforcementPolicy: NewDefaultKeepaliveEnforcementPolicy(),
	}
}

// KeepaliveServerParameters allow configuration of the keepalive.ServerParameters.
// The same default values as keepalive.ServerParameters are applicable and get applied by the server.
// See https://godoc.org/google.golang.org/grpc/keepalive#ServerParameters for details.
type KeepaliveServerParameters struct {
	MaxConnectionIdle     time.Duration `mapstructure:"max_connection_idle"`
	MaxConnectionAge      time.Duration `mapstructure:"max_connection_age"`
	MaxConnectionAgeGrace time.Duration `mapstructure:"max_connection_age_grace"`
	Time                  time.Duration `mapstructure:"time"`
	Timeout               time.Duration `mapstructure:"timeout"`
}

// NewDefaultKeepaliveServerParameters creates and returns a new instance of KeepaliveServerParameters with default settings.
func NewDefaultKeepaliveServerParameters() *KeepaliveServerParameters {
	return &KeepaliveServerParameters{}
}

// KeepaliveEnforcementPolicy allow configuration of the keepalive.EnforcementPolicy.
// The same default values as keepalive.EnforcementPolicy are applicable and get applied by the server.
// See https://godoc.org/google.golang.org/grpc/keepalive#EnforcementPolicy for details.
type KeepaliveEnforcementPolicy struct {
	MinTime             time.Duration `mapstructure:"min_time"`
	PermitWithoutStream bool          `mapstructure:"permit_without_stream"`
}

// NewDefaultKeepaliveEnforcementPolicy creates and returns a new instance of KeepaliveEnforcementPolicy with default settings.
func NewDefaultKeepaliveEnforcementPolicy() *KeepaliveEnforcementPolicy {
	return &KeepaliveEnforcementPolicy{}
}

// ServerConfig defines common settings for a gRPC server configuration.
type ServerConfig struct {
	// Server net.Addr config. For transport only "tcp" and "unix" are valid options.
	NetAddr confignet.AddrConfig `mapstructure:",squash"`

	// Configures the protocol to use TLS.
	// The default value is nil, which will cause the protocol to not use TLS.
	TLSSetting *configtls.ServerConfig `mapstructure:"tls"`

	// MaxRecvMsgSizeMiB sets the maximum size (in MiB) of messages accepted by the server.
	MaxRecvMsgSizeMiB uint64 `mapstructure:"max_recv_msg_size_mib"`

	// MaxConcurrentStreams sets the limit on the number of concurrent streams to each ServerTransport.
	// It has effect only for streaming RPCs.
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams"`

	// ReadBufferSize for gRPC server. See grpc.ReadBufferSize.
	// (https://godoc.org/google.golang.org/grpc#ReadBufferSize).
	ReadBufferSize int `mapstructure:"read_buffer_size"`

	// WriteBufferSize for gRPC server. See grpc.WriteBufferSize.
	// (https://godoc.org/google.golang.org/grpc#WriteBufferSize).
	WriteBufferSize int `mapstructure:"write_buffer_size"`

	// Keepalive anchor for all the settings related to keepalive.
	Keepalive *KeepaliveServerConfig `mapstructure:"keepalive"`

	// Auth for this receiver
	Auth *configauth.Authentication `mapstructure:"auth"`

	// Include propagates the incoming connection's metadata to downstream consumers.
	// Experimental: *NOTE* this option is subject to change or removal in the future.
	IncludeMetadata bool `mapstructure:"include_metadata"`
}

// NewDefaultServerConfig returns a new instance of ServerConfig with default values.
func NewDefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Keepalive: NewDefaultKeepaliveServerConfig(),
		Auth:      configauth.NewDefaultAuthentication(),
	}
}

// sanitizedEndpoint strips the prefix of either http:// or https:// from configgrpc.ClientConfig.Endpoint.
func (gcs *ClientConfig) sanitizedEndpoint() string {
	switch {
	case gcs.isSchemeHTTP():
		return strings.TrimPrefix(gcs.Endpoint, "http://")
	case gcs.isSchemeHTTPS():
		return strings.TrimPrefix(gcs.Endpoint, "https://")
	default:
		return gcs.Endpoint
	}
}

func (gcs *ClientConfig) isSchemeHTTP() bool {
	return strings.HasPrefix(gcs.Endpoint, "http://")
}

func (gcs *ClientConfig) isSchemeHTTPS() bool {
	return strings.HasPrefix(gcs.Endpoint, "https://")
}

// ToClientConn creates a client connection to the given target. By default, it's
// a non-blocking dial (the function won't wait for connections to be
// established, and connecting happens in the background). To make it a blocking
// dial, use grpc.WithBlock() dial option.
func (gcs *ClientConfig) ToClientConn(ctx context.Context, host component.Host, settings component.TelemetrySettings, extraOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts, err := gcs.toDialOptions(ctx, host, settings)
	if err != nil {
		return nil, err
	}
	opts = append(opts, extraOpts...)
	return grpc.NewClient(gcs.sanitizedEndpoint(), opts...)
}

func (gcs *ClientConfig) toDialOptions(ctx context.Context, host component.Host, settings component.TelemetrySettings) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption
	if gcs.Compression.IsCompressed() {
		cp, err := getGRPCCompressionName(gcs.Compression)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor(cp)))
	}

	tlsCfg, err := gcs.TLSSetting.LoadTLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	cred := insecure.NewCredentials()
	if tlsCfg != nil {
		cred = credentials.NewTLS(tlsCfg)
	} else if gcs.isSchemeHTTPS() {
		cred = credentials.NewTLS(&tls.Config{})
	}
	opts = append(opts, grpc.WithTransportCredentials(cred))

	if gcs.ReadBufferSize > 0 {
		opts = append(opts, grpc.WithReadBufferSize(gcs.ReadBufferSize))
	}

	if gcs.WriteBufferSize > 0 {
		opts = append(opts, grpc.WithWriteBufferSize(gcs.WriteBufferSize))
	}

	if gcs.Keepalive != nil {
		keepAliveOption := grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                gcs.Keepalive.Time,
			Timeout:             gcs.Keepalive.Timeout,
			PermitWithoutStream: gcs.Keepalive.PermitWithoutStream,
		})
		opts = append(opts, keepAliveOption)
	}

	if gcs.Auth != nil {
		if host.GetExtensions() == nil {
			return nil, errors.New("no extensions configuration available")
		}

		grpcAuthenticator, cerr := gcs.Auth.GetClientAuthenticatorContext(ctx, host.GetExtensions())
		if cerr != nil {
			return nil, cerr
		}

		perRPCCredentials, perr := grpcAuthenticator.PerRPCCredentials()
		if perr != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithPerRPCCredentials(perRPCCredentials))
	}

	if gcs.BalancerName != "" {
		valid := validateBalancerName(gcs.BalancerName)
		if !valid {
			return nil, fmt.Errorf("invalid balancer_name: %s", gcs.BalancerName)
		}
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, gcs.BalancerName)))
	}

	if gcs.Authority != "" {
		opts = append(opts, grpc.WithAuthority(gcs.Authority))
	}

	otelOpts := []otelgrpc.Option{
		otelgrpc.WithTracerProvider(settings.TracerProvider),
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
	}
	if settings.MetricsLevel >= configtelemetry.LevelDetailed {
		otelOpts = append(otelOpts, otelgrpc.WithMeterProvider(settings.MeterProvider))
	}

	// Enable OpenTelemetry observability plugin.
	opts = append(opts, grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelOpts...)))

	return opts, nil
}

func validateBalancerName(balancerName string) bool {
	return balancer.Get(balancerName) != nil
}

// ToServer returns a grpc.Server for the configuration
func (gss *ServerConfig) ToServer(_ context.Context, host component.Host, settings component.TelemetrySettings, extraOpts ...grpc.ServerOption) (*grpc.Server, error) {
	opts, err := gss.toServerOption(host, settings)
	if err != nil {
		return nil, err
	}
	opts = append(opts, extraOpts...)
	return grpc.NewServer(opts...), nil
}

func (gss *ServerConfig) toServerOption(host component.Host, settings component.TelemetrySettings) ([]grpc.ServerOption, error) {
	switch gss.NetAddr.Transport {
	case confignet.TransportTypeTCP, confignet.TransportTypeTCP4, confignet.TransportTypeTCP6, confignet.TransportTypeUDP, confignet.TransportTypeUDP4, confignet.TransportTypeUDP6:
		internal.WarnOnUnspecifiedHost(settings.Logger, gss.NetAddr.Endpoint)
	}

	var opts []grpc.ServerOption

	if gss.TLSSetting != nil {
		tlsCfg, err := gss.TLSSetting.LoadTLSConfig(context.Background())
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsCfg)))
	}

	if gss.MaxRecvMsgSizeMiB > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(int(gss.MaxRecvMsgSizeMiB*1024*1024)))
	}

	if gss.MaxConcurrentStreams > 0 {
		opts = append(opts, grpc.MaxConcurrentStreams(gss.MaxConcurrentStreams))
	}

	if gss.ReadBufferSize > 0 {
		opts = append(opts, grpc.ReadBufferSize(gss.ReadBufferSize))
	}

	if gss.WriteBufferSize > 0 {
		opts = append(opts, grpc.WriteBufferSize(gss.WriteBufferSize))
	}

	// The default values referenced in the GRPC docs are set within the server, so this code doesn't need
	// to apply them over zero/nil values before passing these as grpc.ServerOptions.
	// The following shows the server code for applying default grpc.ServerOptions.
	// https://github.com/grpc/grpc-go/blob/120728e1f775e40a2a764341939b78d666b08260/internal/transport/http2_server.go#L184-L200
	if gss.Keepalive != nil {
		if gss.Keepalive.ServerParameters != nil {
			svrParams := gss.Keepalive.ServerParameters
			opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle:     svrParams.MaxConnectionIdle,
				MaxConnectionAge:      svrParams.MaxConnectionAge,
				MaxConnectionAgeGrace: svrParams.MaxConnectionAgeGrace,
				Time:                  svrParams.Time,
				Timeout:               svrParams.Timeout,
			}))
		}
		// The default values referenced in the GRPC are set within the server, so this code doesn't need
		// to apply them over zero/nil values before passing these as grpc.ServerOptions.
		// The following shows the server code for applying default grpc.ServerOptions.
		// https://github.com/grpc/grpc-go/blob/120728e1f775e40a2a764341939b78d666b08260/internal/transport/http2_server.go#L202-L205
		if gss.Keepalive.EnforcementPolicy != nil {
			enfPol := gss.Keepalive.EnforcementPolicy
			opts = append(opts, grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             enfPol.MinTime,
				PermitWithoutStream: enfPol.PermitWithoutStream,
			}))
		}
	}

	var uInterceptors []grpc.UnaryServerInterceptor
	var sInterceptors []grpc.StreamServerInterceptor

	if gss.Auth != nil {
		authenticator, err := gss.Auth.GetServerAuthenticatorContext(context.Background(), host.GetExtensions())
		if err != nil {
			return nil, err
		}

		uInterceptors = append(uInterceptors, func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			return authUnaryServerInterceptor(ctx, req, info, handler, authenticator)
		})
		sInterceptors = append(sInterceptors, func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return authStreamServerInterceptor(srv, ss, info, handler, authenticator)
		})
	}

	otelOpts := []otelgrpc.Option{
		otelgrpc.WithTracerProvider(settings.TracerProvider),
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
	}
	if settings.MetricsLevel >= configtelemetry.LevelDetailed {
		otelOpts = append(otelOpts, otelgrpc.WithMeterProvider(settings.MeterProvider))
	}

	// Enable OpenTelemetry observability plugin.

	uInterceptors = append(uInterceptors, enhanceWithClientInformation(gss.IncludeMetadata))
	sInterceptors = append(sInterceptors, enhanceStreamWithClientInformation(gss.IncludeMetadata))

	opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler(otelOpts...)), grpc.ChainUnaryInterceptor(uInterceptors...), grpc.ChainStreamInterceptor(sInterceptors...))

	return opts, nil
}

// getGRPCCompressionName returns compression name registered in grpc.
func getGRPCCompressionName(compressionType configcompression.Type) (string, error) {
	switch compressionType {
	case configcompression.TypeGzip:
		return gzip.Name, nil
	case configcompression.TypeSnappy:
		return snappy.Name, nil
	case configcompression.TypeZstd:
		return zstd.Name, nil
	default:
		return "", fmt.Errorf("unsupported compression type %q", compressionType)
	}
}

// enhanceWithClientInformation intercepts the incoming RPC, replacing the incoming context with one that includes
// a client.Info, potentially with the peer's address.
func enhanceWithClientInformation(includeMetadata bool) func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(contextWithClient(ctx, includeMetadata), req)
	}
}

func enhanceStreamWithClientInformation(includeMetadata bool) func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, wrapServerStream(contextWithClient(ss.Context(), includeMetadata), ss))
	}
}

// contextWithClient attempts to add the peer address to the client.Info from the context. When no
// client.Info exists in the context, one is created.
func contextWithClient(ctx context.Context, includeMetadata bool) context.Context {
	cl := client.FromContext(ctx)
	if p, ok := peer.FromContext(ctx); ok {
		cl.Addr = p.Addr
	}
	if includeMetadata {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			copiedMD := md.Copy()
			if len(md[client.MetadataHostName]) == 0 && len(md[":authority"]) > 0 {
				copiedMD[client.MetadataHostName] = md[":authority"]
			}
			cl.Metadata = client.NewMetadata(copiedMD)
		}
	}
	return client.NewContext(ctx, cl)
}

func authUnaryServerInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler, server auth.Server) (any, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMetadataNotFound
	}

	ctx, err := server.Authenticate(ctx, headers)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func authStreamServerInterceptor(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler, server auth.Server) error {
	ctx := stream.Context()
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errMetadataNotFound
	}

	ctx, err := server.Authenticate(ctx, headers)
	if err != nil {
		return err
	}

	return handler(srv, wrapServerStream(ctx, stream))
}
