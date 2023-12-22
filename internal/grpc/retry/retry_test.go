// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package retry

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	retriableErrors = []codes.Code{codes.Unavailable, codes.DataLoss}
	noSleep         = 0 * time.Second
	retryTimeout    = 50 * time.Millisecond
)

type failingService struct {
	testpb.TestServiceServer
	mu sync.Mutex

	reqCounter uint
	reqModulo  uint
	reqSleep   time.Duration
	reqError   codes.Code
}

func (s *failingService) resetFailingConfiguration(modulo uint, errorCode codes.Code, sleepTime time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reqCounter = 0
	s.reqModulo = modulo
	s.reqError = errorCode
	s.reqSleep = sleepTime
}

func (s *failingService) requestCount() uint {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reqCounter
}

func (s *failingService) maybeFailRequest() error {
	s.mu.Lock()
	s.reqCounter += 1
	reqModulo := s.reqModulo
	reqCounter := s.reqCounter
	reqSleep := s.reqSleep
	reqError := s.reqError
	s.mu.Unlock()
	if (reqModulo > 0) && (reqCounter%reqModulo == 0) {
		return nil
	}
	time.Sleep(reqSleep)
	return status.Error(reqError, "maybeFailRequest: failing it")
}

func (s *failingService) Ping(ctx context.Context, ping *testpb.PingRequest) (*testpb.PingResponse, error) {
	if err := s.maybeFailRequest(); err != nil {
		return nil, err
	}
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *failingService) PingList(ping *testpb.PingListRequest, stream testpb.TestService_PingListServer) error {
	if err := s.maybeFailRequest(); err != nil {
		return err
	}
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *failingService) PingStream(stream testpb.TestService_PingStreamServer) error {
	if err := s.maybeFailRequest(); err != nil {
		return err
	}
	return s.TestServiceServer.PingStream(stream)
}

func TestRetrySuite(t *testing.T) {
	service := &failingService{
		TestServiceServer: &testpb.TestPingService{},
	}
	unaryInterceptor := UnaryClientInterceptor(
		WithCodes(retriableErrors...),
		WithMax(3),
		WithBackoff(BackoffLinear(retryTimeout)),
	)
	streamInterceptor := StreamClientInterceptor(
		WithCodes(retriableErrors...),
		WithMax(3),
		WithBackoff(BackoffLinear(retryTimeout)),
	)
	s := &RetrySuite{
		srv: service,
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: service,
			ClientOpts: []grpc.DialOption{
				grpc.WithStreamInterceptor(streamInterceptor),
				grpc.WithUnaryInterceptor(unaryInterceptor),
			},
		},
	}
	suite.Run(t, s)
}

type RetrySuite struct {
	*testpb.InterceptorTestSuite
	srv *failingService
}

func (s *RetrySuite) SetupTest() {
	s.srv.resetFailingConfiguration( /* don't fail */ 0, codes.OK, noSleep)
}

func (s *RetrySuite) TestUnary_FailsOnNonRetriableError() {
	s.srv.resetFailingConfiguration(5, codes.Internal, noSleep)
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	require.Error(s.T(), err, "error must occur from the failing service")
	require.Equal(s.T(), codes.Internal, status.Code(err), "failure code must come from retrier")
	require.EqualValues(s.T(), 1, s.srv.requestCount(), "one request should have been made")
}

func (s *RetrySuite) TestUnary_FailsOnNonRetriableContextError() {
	s.srv.resetFailingConfiguration(5, codes.Canceled, noSleep)
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	require.Error(s.T(), err, "error must occur from the failing service")
	require.Equal(s.T(), codes.Canceled, status.Code(err), "failure code must come from retrier")
	require.EqualValues(s.T(), 1, s.srv.requestCount(), "one request should have been made")
}

func (s *RetrySuite) TestCallOptionsDontPanicWithoutInterceptor() {
	// Fix for https://github.com/grpc-ecosystem/go-grpc-middleware/issues/37
	// If this code doesn't panic, that's good.
	s.srv.resetFailingConfiguration(100, codes.DataLoss, noSleep) // doesn't matter all requests should fail
	nonMiddlewareClient := s.NewClient()
	_, err := nonMiddlewareClient.Ping(s.SimpleCtx(), testpb.GoodPing,
		WithMax(5),
		WithBackoff(BackoffLinear(1*time.Millisecond)),
		WithCodes(codes.DataLoss),
		WithPerRetryTimeout(1*time.Millisecond),
	)
	require.Error(s.T(), err)
}

func (s *RetrySuite) TestServerStream_FailsOnNonRetriableError() {
	s.srv.resetFailingConfiguration(5, codes.Internal, noSleep)
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	_, err = stream.Recv()
	require.Error(s.T(), err, "error must occur from the failing service")
	require.Equal(s.T(), codes.Internal, status.Code(err), "failure code must come from retrier")
}

func (s *RetrySuite) TestUnary_SucceedsOnRetriableError() {
	s.srv.resetFailingConfiguration(3, codes.DataLoss, noSleep) // see retriable_errors
	out, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing)
	require.NoError(s.T(), err, "the third invocation should succeed")
	require.NotNil(s.T(), out, "Pong must be not nil")
	require.EqualValues(s.T(), 3, s.srv.requestCount(), "three requests should have been made")
}

func (s *RetrySuite) TestUnary_OverrideFromDialOpts() {
	s.srv.resetFailingConfiguration(5, codes.ResourceExhausted, noSleep) // default is 3 and retriable_errors
	out, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing, WithCodes(codes.ResourceExhausted), WithMax(5))
	require.NoError(s.T(), err, "the fifth invocation should succeed")
	require.NotNil(s.T(), out, "Pong must be not nil")
	require.EqualValues(s.T(), 5, s.srv.requestCount(), "five requests should have been made")
}

func (s *RetrySuite) TestUnary_OnRetryCallbackCalled() {
	retryCallbackCount := 0

	s.srv.resetFailingConfiguration(3, codes.Unavailable, noSleep) // see retriable_errors
	out, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing,
		WithOnRetryCallback(func(ctx context.Context, attempt uint, err error) {
			retryCallbackCount++
		}),
	)

	require.NoError(s.T(), err, "the third invocation should succeed")
	require.NotNil(s.T(), out, "Pong must be not nil")
	require.EqualValues(s.T(), 2, retryCallbackCount, "two retry callbacks should be called")
}

func (s *RetrySuite) TestServerStream_SucceedsOnRetriableError() {
	s.srv.resetFailingConfiguration(3, codes.DataLoss, noSleep) // see retriable_errors
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)
	require.NoError(s.T(), err, "establishing the connection must always succeed")
	s.assertPingListWasCorrect(stream)
	require.EqualValues(s.T(), 3, s.srv.requestCount(), "three requests should have been made")
}

func (s *RetrySuite) TestServerStream_OverrideFromContext() {
	s.srv.resetFailingConfiguration(5, codes.ResourceExhausted, noSleep) // default is 3 and retriable_errors
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList, WithCodes(codes.ResourceExhausted), WithMax(5))
	require.NoError(s.T(), err, "establishing the connection must always succeed")
	s.assertPingListWasCorrect(stream)
	require.EqualValues(s.T(), 5, s.srv.requestCount(), "three requests should have been made")
}

func (s *RetrySuite) TestServerStream_OnRetryCallbackCalled() {
	retryCallbackCount := 0

	s.srv.resetFailingConfiguration(3, codes.Unavailable, noSleep) // see retriable_errors
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList,
		WithOnRetryCallback(func(ctx context.Context, attempt uint, err error) {
			retryCallbackCount++
		}),
	)

	require.NoError(s.T(), err, "establishing the connection must always succeed")
	s.assertPingListWasCorrect(stream)
	require.EqualValues(s.T(), 2, retryCallbackCount, "two retry callbacks should be called")
}

func (s *RetrySuite) TestServerStream_CallFailsOnOutOfRetries() {
	restarted := s.RestartServer(3 * retryTimeout)
	_, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList)

	require.Error(s.T(), err, "establishing the connection should not succeed")
	assert.Equal(s.T(), codes.Unavailable, status.Code(err))

	<-restarted
}

func (s *RetrySuite) TestServerStream_CallFailsOnDeadlineExceeded() {
	restarted := s.RestartServer(3 * retryTimeout)
	ctx, cancel := context.WithTimeout(context.TODO(), retryTimeout)
	defer cancel()
	_, err := s.Client.PingList(ctx, testpb.GoodPingList)

	require.Error(s.T(), err, "establishing the connection should not succeed")
	assert.Equal(s.T(), codes.DeadlineExceeded, status.Code(err))

	<-restarted
}

func (s *RetrySuite) TestServerStream_CallRetrySucceeds() {
	restarted := s.RestartServer(retryTimeout)

	_, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList,
		WithMax(40),
	)

	assert.NoError(s.T(), err, "establishing the connection should succeed")
	<-restarted
}

func (s *RetrySuite) assertPingListWasCorrect(stream testpb.TestService_PingListClient) {
	count := 0
	for {
		pong, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "no errors during receive on client side")
		require.NotNil(s.T(), pong, "received values must not be nil")
		require.Equal(s.T(), testpb.GoodPingList.Value, pong.Value, "the returned pong contained the outgoing ping")
		count += 1
	}
	require.EqualValues(s.T(), testpb.ListResponseCount, count, "should have received all ping items")
}

type trackedInterceptor struct {
	called int
}

func (ti *trackedInterceptor) UnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ti.called++
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (ti *trackedInterceptor) StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ti.called++
	return streamer(ctx, desc, cc, method, opts...)
}

func TestChainedRetrySuite(t *testing.T) {
	service := &failingService{
		TestServiceServer: &testpb.TestPingService{},
	}
	preRetryInterceptor := &trackedInterceptor{}
	postRetryInterceptor := &trackedInterceptor{}
	s := &ChainedRetrySuite{
		srv:                  service,
		preRetryInterceptor:  preRetryInterceptor,
		postRetryInterceptor: postRetryInterceptor,
		InterceptorTestSuite: &testpb.InterceptorTestSuite{
			TestService: service,
			ClientOpts: []grpc.DialOption{
				grpc.WithChainUnaryInterceptor(
					preRetryInterceptor.UnaryClientInterceptor,
					UnaryClientInterceptor(),
					postRetryInterceptor.UnaryClientInterceptor,
				),
				grpc.WithChainStreamInterceptor(
					preRetryInterceptor.StreamClientInterceptor,
					StreamClientInterceptor(),
					postRetryInterceptor.StreamClientInterceptor,
				),
			},
		},
	}
	suite.Run(t, s)
}

type ChainedRetrySuite struct {
	*testpb.InterceptorTestSuite
	srv                  *failingService
	preRetryInterceptor  *trackedInterceptor
	postRetryInterceptor *trackedInterceptor
}

func (s *ChainedRetrySuite) SetupTest() {
	s.srv.resetFailingConfiguration( /* don't fail */ 0, codes.OK, noSleep)
	s.preRetryInterceptor.called = 0
	s.postRetryInterceptor.called = 0
}

func (s *ChainedRetrySuite) TestUnaryWithChainedInterceptors_NoFailure() {
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing, WithMax(2))
	require.NoError(s.T(), err, "the invocation should succeed")
	require.EqualValues(s.T(), 1, s.srv.requestCount(), "one request should have been made")
	require.EqualValues(s.T(), 1, s.preRetryInterceptor.called, "pre-retry interceptor should be called once")
	require.EqualValues(s.T(), 1, s.postRetryInterceptor.called, "post-retry interceptor should be called once")
}

func (s *ChainedRetrySuite) TestUnaryWithChainedInterceptors_WithRetry() {
	s.srv.resetFailingConfiguration(2, codes.Unavailable, noSleep)
	_, err := s.Client.Ping(s.SimpleCtx(), testpb.GoodPing, WithMax(2))
	require.NoError(s.T(), err, "the second invocation should succeed")
	require.EqualValues(s.T(), 2, s.srv.requestCount(), "two requests should have been made")
	require.EqualValues(s.T(), 1, s.preRetryInterceptor.called, "pre-retry interceptor should be called once")
	require.EqualValues(s.T(), 2, s.postRetryInterceptor.called, "post-retry interceptor should be called twice")
}

func (s *ChainedRetrySuite) TestStreamWithChainedInterceptors_NoFailure() {
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList, WithMax(2))
	require.NoError(s.T(), err, "the invocation should succeed")
	_, err = stream.Recv()
	require.NoError(s.T(), err, "the Recv should succeed")
	require.EqualValues(s.T(), 1, s.srv.requestCount(), "one request should have been made")
	require.EqualValues(s.T(), 1, s.preRetryInterceptor.called, "pre-retry interceptor should be called once")
	require.EqualValues(s.T(), 1, s.postRetryInterceptor.called, "post-retry interceptor should be called once")
}

func (s *ChainedRetrySuite) TestStreamWithChainedInterceptors_WithRetry() {
	s.srv.resetFailingConfiguration(2, codes.Unavailable, noSleep)
	stream, err := s.Client.PingList(s.SimpleCtx(), testpb.GoodPingList, WithMax(2))
	require.NoError(s.T(), err, "the second invocation should succeed")
	_, err = stream.Recv()
	require.NoError(s.T(), err, "the Recv should succeed")
	require.EqualValues(s.T(), 2, s.srv.requestCount(), "two requests should have been made")
	require.EqualValues(s.T(), 1, s.preRetryInterceptor.called, "pre-retry interceptor should be called once")
	require.EqualValues(s.T(), 2, s.postRetryInterceptor.called, "post-retry interceptor should be called twice")
}

func TestJitterUp(t *testing.T) {
	// Arguments to jitterup.
	duration := 10 * time.Second
	variance := 0.10

	// Bound to check.
	max := 11000 * time.Millisecond
	min := 9000 * time.Millisecond
	high := scaleDuration(max, 0.98)
	low := scaleDuration(min, 1.02)

	highCount := 0
	lowCount := 0

	for i := 0; i < 1000; i++ {
		out := jitterUp(duration, variance)
		assert.True(t, out <= max, "value %s must be <= %s", out, max)
		assert.True(t, out >= min, "value %s must be >= %s", out, min)

		if out > high {
			highCount++
		}
		if out < low {
			lowCount++
		}
	}

	assert.True(t, highCount != 0, "at least one sample should reach to >%s", high)
	assert.True(t, lowCount != 0, "at least one sample should to <%s", low)
}
