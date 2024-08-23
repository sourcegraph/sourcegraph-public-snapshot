// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterhelper // import "go.opentelemetry.io/collector/exporter/exporterhelper"

import (
	"context"
	"fmt"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterbatcher"
	"go.opentelemetry.io/collector/exporter/exporterqueue"
)

// requestSender is an abstraction of a sender for a request independent of the type of the data (traces, metrics, logs).
type requestSender interface {
	component.Component
	send(context.Context, Request) error
	setNextSender(nextSender requestSender)
}

type baseRequestSender struct {
	component.StartFunc
	component.ShutdownFunc
	nextSender requestSender
}

var _ requestSender = (*baseRequestSender)(nil)

func (b *baseRequestSender) send(ctx context.Context, req Request) error {
	return b.nextSender.send(ctx, req)
}

func (b *baseRequestSender) setNextSender(nextSender requestSender) {
	b.nextSender = nextSender
}

type obsrepSenderFactory func(obsrep *ObsReport) requestSender

// Option apply changes to baseExporter.
type Option func(*baseExporter) error

// WithStart overrides the default Start function for an exporter.
// The default start function does nothing and always returns nil.
func WithStart(start component.StartFunc) Option {
	return func(o *baseExporter) error {
		o.StartFunc = start
		return nil
	}
}

// WithShutdown overrides the default Shutdown function for an exporter.
// The default shutdown function does nothing and always returns nil.
func WithShutdown(shutdown component.ShutdownFunc) Option {
	return func(o *baseExporter) error {
		o.ShutdownFunc = shutdown
		return nil
	}
}

// WithTimeout overrides the default TimeoutSettings for an exporter.
// The default TimeoutSettings is 5 seconds.
func WithTimeout(timeoutSettings TimeoutSettings) Option {
	return func(o *baseExporter) error {
		o.timeoutSender.cfg = timeoutSettings
		return nil
	}
}

// WithRetry overrides the default configretry.BackOffConfig for an exporter.
// The default configretry.BackOffConfig is to disable retries.
func WithRetry(config configretry.BackOffConfig) Option {
	return func(o *baseExporter) error {
		if !config.Enabled {
			o.exportFailureMessage += " Try enabling retry_on_failure config option to retry on retryable errors."
			return nil
		}
		o.retrySender = newRetrySender(config, o.set)
		return nil
	}
}

// WithQueue overrides the default QueueSettings for an exporter.
// The default QueueSettings is to disable queueing.
// This option cannot be used with the new exporter helpers New[Traces|Metrics|Logs]RequestExporter.
func WithQueue(config QueueSettings) Option {
	return func(o *baseExporter) error {
		if o.marshaler == nil || o.unmarshaler == nil {
			return fmt.Errorf("WithQueue option is not available for the new request exporters, use WithRequestQueue instead")
		}
		if !config.Enabled {
			o.exportFailureMessage += " Try enabling sending_queue to survive temporary failures."
			return nil
		}
		qf := exporterqueue.NewPersistentQueueFactory[Request](config.StorageID, exporterqueue.PersistentQueueSettings[Request]{
			Marshaler:   o.marshaler,
			Unmarshaler: o.unmarshaler,
		})
		q := qf(context.Background(), exporterqueue.Settings{
			DataType:         o.signal,
			ExporterSettings: o.set,
		}, exporterqueue.Config{
			Enabled:      config.Enabled,
			NumConsumers: config.NumConsumers,
			QueueSize:    config.QueueSize,
		})
		o.queueSender = newQueueSender(q, o.set, config.NumConsumers, o.exportFailureMessage, o.obsrep.telemetryBuilder)
		return nil
	}
}

// WithRequestQueue enables queueing for an exporter.
// This option should be used with the new exporter helpers New[Traces|Metrics|Logs]RequestExporter.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
func WithRequestQueue(cfg exporterqueue.Config, queueFactory exporterqueue.Factory[Request]) Option {
	return func(o *baseExporter) error {
		if o.marshaler != nil || o.unmarshaler != nil {
			return fmt.Errorf("WithRequestQueue option must be used with the new request exporters only, use WithQueue instead")
		}
		if !cfg.Enabled {
			o.exportFailureMessage += " Try enabling sending_queue to survive temporary failures."
			return nil
		}
		set := exporterqueue.Settings{
			DataType:         o.signal,
			ExporterSettings: o.set,
		}
		o.queueSender = newQueueSender(queueFactory(context.Background(), set, cfg), o.set, cfg.NumConsumers, o.exportFailureMessage, o.obsrep.telemetryBuilder)
		return nil
	}
}

// WithCapabilities overrides the default Capabilities() function for a Consumer.
// The default is non-mutable data.
// TODO: Verify if we can change the default to be mutable as we do for processors.
func WithCapabilities(capabilities consumer.Capabilities) Option {
	return func(o *baseExporter) error {
		o.consumerOptions = append(o.consumerOptions, consumer.WithCapabilities(capabilities))
		return nil
	}
}

// BatcherOption apply changes to batcher sender.
type BatcherOption func(*batchSender) error

// WithRequestBatchFuncs sets the functions for merging and splitting batches for an exporter built for custom request types.
func WithRequestBatchFuncs(mf exporterbatcher.BatchMergeFunc[Request], msf exporterbatcher.BatchMergeSplitFunc[Request]) BatcherOption {
	return func(bs *batchSender) error {
		if mf == nil || msf == nil {
			return fmt.Errorf("WithRequestBatchFuncs must be provided with non-nil functions")
		}
		if bs.mergeFunc != nil || bs.mergeSplitFunc != nil {
			return fmt.Errorf("WithRequestBatchFuncs can only be used once with request-based exporters")
		}
		bs.mergeFunc = mf
		bs.mergeSplitFunc = msf
		return nil
	}
}

// WithBatcher enables batching for an exporter based on custom request types.
// For now, it can be used only with the New[Traces|Metrics|Logs]RequestExporter exporter helpers and
// WithRequestBatchFuncs provided.
// This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
func WithBatcher(cfg exporterbatcher.Config, opts ...BatcherOption) Option {
	return func(o *baseExporter) error {
		if !cfg.Enabled {
			return nil
		}

		bs := newBatchSender(cfg, o.set, o.batchMergeFunc, o.batchMergeSplitfunc)
		for _, opt := range opts {
			if err := opt(bs); err != nil {
				return err
			}
		}
		if bs.mergeFunc == nil || bs.mergeSplitFunc == nil {
			return fmt.Errorf("WithRequestBatchFuncs must be provided for the batcher applied to the request-based exporters")
		}
		o.batchSender = bs
		return nil
	}
}

// withMarshaler is used to set the request marshaler for the new exporter helper.
// It must be provided as the first option when creating a new exporter helper.
func withMarshaler(marshaler exporterqueue.Marshaler[Request]) Option {
	return func(o *baseExporter) error {
		o.marshaler = marshaler
		return nil
	}
}

// withUnmarshaler is used to set the request unmarshaler for the new exporter helper.
// It must be provided as the first option when creating a new exporter helper.
func withUnmarshaler(unmarshaler exporterqueue.Unmarshaler[Request]) Option {
	return func(o *baseExporter) error {
		o.unmarshaler = unmarshaler
		return nil
	}
}

// withBatchFuncs is used to set the functions for merging and splitting batches for OLTP-based exporters.
// It must be provided as the first option when creating a new exporter helper.
func withBatchFuncs(mf exporterbatcher.BatchMergeFunc[Request], msf exporterbatcher.BatchMergeSplitFunc[Request]) Option {
	return func(o *baseExporter) error {
		o.batchMergeFunc = mf
		o.batchMergeSplitfunc = msf
		return nil
	}
}

// baseExporter contains common fields between different exporter types.
type baseExporter struct {
	component.StartFunc
	component.ShutdownFunc

	signal component.DataType

	batchMergeFunc      exporterbatcher.BatchMergeFunc[Request]
	batchMergeSplitfunc exporterbatcher.BatchMergeSplitFunc[Request]

	marshaler   exporterqueue.Marshaler[Request]
	unmarshaler exporterqueue.Unmarshaler[Request]

	set    exporter.Settings
	obsrep *ObsReport

	// Message for the user to be added with an export failure message.
	exportFailureMessage string

	// Chain of senders that the exporter helper applies before passing the data to the actual exporter.
	// The data is handled by each sender in the respective order starting from the queueSender.
	// Most of the senders are optional, and initialized with a no-op path-through sender.
	batchSender   requestSender
	queueSender   requestSender
	obsrepSender  requestSender
	retrySender   requestSender
	timeoutSender *timeoutSender // timeoutSender is always initialized.

	consumerOptions []consumer.Option
}

func newBaseExporter(set exporter.Settings, signal component.DataType, osf obsrepSenderFactory, options ...Option) (*baseExporter, error) {
	obsReport, err := NewObsReport(ObsReportSettings{ExporterID: set.ID, ExporterCreateSettings: set})
	if err != nil {
		return nil, err
	}

	be := &baseExporter{
		signal: signal,

		batchSender:   &baseRequestSender{},
		queueSender:   &baseRequestSender{},
		obsrepSender:  osf(obsReport),
		retrySender:   &baseRequestSender{},
		timeoutSender: &timeoutSender{cfg: NewDefaultTimeoutSettings()},

		set:    set,
		obsrep: obsReport,
	}

	for _, op := range options {
		err = multierr.Append(err, op(be))
	}
	if err != nil {
		return nil, err
	}

	be.connectSenders()

	if bs, ok := be.batchSender.(*batchSender); ok {
		// If queue sender is enabled assign to the batch sender the same number of workers.
		if qs, ok := be.queueSender.(*queueSender); ok {
			bs.concurrencyLimit = int64(qs.numConsumers)
		}
		// Batcher sender mutates the data.
		be.consumerOptions = append(be.consumerOptions, consumer.WithCapabilities(consumer.Capabilities{MutatesData: true}))
	}

	return be, nil
}

// send sends the request using the first sender in the chain.
func (be *baseExporter) send(ctx context.Context, req Request) error {
	err := be.queueSender.send(ctx, req)
	if err != nil {
		be.set.Logger.Error("Exporting failed. Rejecting data."+be.exportFailureMessage,
			zap.Error(err), zap.Int("rejected_items", req.ItemsCount()))
	}
	return err
}

// connectSenders connects the senders in the predefined order.
func (be *baseExporter) connectSenders() {
	be.queueSender.setNextSender(be.batchSender)
	be.batchSender.setNextSender(be.obsrepSender)
	be.obsrepSender.setNextSender(be.retrySender)
	be.retrySender.setNextSender(be.timeoutSender)
}

func (be *baseExporter) Start(ctx context.Context, host component.Host) error {
	// First start the wrapped exporter.
	if err := be.StartFunc.Start(ctx, host); err != nil {
		return err
	}

	// If no error then start the batchSender.
	if err := be.batchSender.Start(ctx, host); err != nil {
		return err
	}

	// Last start the queueSender.
	return be.queueSender.Start(ctx, host)
}

func (be *baseExporter) Shutdown(ctx context.Context) error {
	return multierr.Combine(
		// First shutdown the retry sender, so the queue sender can flush the queue without retries.
		be.retrySender.Shutdown(ctx),
		// Then shutdown the batch sender
		be.batchSender.Shutdown(ctx),
		// Then shutdown the queue sender.
		be.queueSender.Shutdown(ctx),
		// Last shutdown the wrapped exporter itself.
		be.ShutdownFunc.Shutdown(ctx))
}
