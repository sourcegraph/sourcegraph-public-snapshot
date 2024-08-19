// Copyright 2024 Google LLC
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

// Code generated by protoc-gen-go_gapic. DO NOT EDIT.

package monitoring

import (
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/iterator"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
)

// AlertPolicyIterator manages a stream of *monitoringpb.AlertPolicy.
type AlertPolicyIterator struct {
	items    []*monitoringpb.AlertPolicy
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.AlertPolicy, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *AlertPolicyIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *AlertPolicyIterator) Next() (*monitoringpb.AlertPolicy, error) {
	var item *monitoringpb.AlertPolicy
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *AlertPolicyIterator) bufLen() int {
	return len(it.items)
}

func (it *AlertPolicyIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// GroupIterator manages a stream of *monitoringpb.Group.
type GroupIterator struct {
	items    []*monitoringpb.Group
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.Group, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *GroupIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *GroupIterator) Next() (*monitoringpb.Group, error) {
	var item *monitoringpb.Group
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *GroupIterator) bufLen() int {
	return len(it.items)
}

func (it *GroupIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// MetricDescriptorIterator manages a stream of *metricpb.MetricDescriptor.
type MetricDescriptorIterator struct {
	items    []*metricpb.MetricDescriptor
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*metricpb.MetricDescriptor, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *MetricDescriptorIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *MetricDescriptorIterator) Next() (*metricpb.MetricDescriptor, error) {
	var item *metricpb.MetricDescriptor
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *MetricDescriptorIterator) bufLen() int {
	return len(it.items)
}

func (it *MetricDescriptorIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// MonitoredResourceDescriptorIterator manages a stream of *monitoredrespb.MonitoredResourceDescriptor.
type MonitoredResourceDescriptorIterator struct {
	items    []*monitoredrespb.MonitoredResourceDescriptor
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoredrespb.MonitoredResourceDescriptor, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *MonitoredResourceDescriptorIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *MonitoredResourceDescriptorIterator) Next() (*monitoredrespb.MonitoredResourceDescriptor, error) {
	var item *monitoredrespb.MonitoredResourceDescriptor
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *MonitoredResourceDescriptorIterator) bufLen() int {
	return len(it.items)
}

func (it *MonitoredResourceDescriptorIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// MonitoredResourceIterator manages a stream of *monitoredrespb.MonitoredResource.
type MonitoredResourceIterator struct {
	items    []*monitoredrespb.MonitoredResource
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoredrespb.MonitoredResource, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *MonitoredResourceIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *MonitoredResourceIterator) Next() (*monitoredrespb.MonitoredResource, error) {
	var item *monitoredrespb.MonitoredResource
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *MonitoredResourceIterator) bufLen() int {
	return len(it.items)
}

func (it *MonitoredResourceIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// NotificationChannelDescriptorIterator manages a stream of *monitoringpb.NotificationChannelDescriptor.
type NotificationChannelDescriptorIterator struct {
	items    []*monitoringpb.NotificationChannelDescriptor
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.NotificationChannelDescriptor, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *NotificationChannelDescriptorIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *NotificationChannelDescriptorIterator) Next() (*monitoringpb.NotificationChannelDescriptor, error) {
	var item *monitoringpb.NotificationChannelDescriptor
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *NotificationChannelDescriptorIterator) bufLen() int {
	return len(it.items)
}

func (it *NotificationChannelDescriptorIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// NotificationChannelIterator manages a stream of *monitoringpb.NotificationChannel.
type NotificationChannelIterator struct {
	items    []*monitoringpb.NotificationChannel
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.NotificationChannel, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *NotificationChannelIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *NotificationChannelIterator) Next() (*monitoringpb.NotificationChannel, error) {
	var item *monitoringpb.NotificationChannel
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *NotificationChannelIterator) bufLen() int {
	return len(it.items)
}

func (it *NotificationChannelIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// ServiceIterator manages a stream of *monitoringpb.Service.
type ServiceIterator struct {
	items    []*monitoringpb.Service
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.Service, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *ServiceIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *ServiceIterator) Next() (*monitoringpb.Service, error) {
	var item *monitoringpb.Service
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *ServiceIterator) bufLen() int {
	return len(it.items)
}

func (it *ServiceIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// ServiceLevelObjectiveIterator manages a stream of *monitoringpb.ServiceLevelObjective.
type ServiceLevelObjectiveIterator struct {
	items    []*monitoringpb.ServiceLevelObjective
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.ServiceLevelObjective, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *ServiceLevelObjectiveIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *ServiceLevelObjectiveIterator) Next() (*monitoringpb.ServiceLevelObjective, error) {
	var item *monitoringpb.ServiceLevelObjective
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *ServiceLevelObjectiveIterator) bufLen() int {
	return len(it.items)
}

func (it *ServiceLevelObjectiveIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// SnoozeIterator manages a stream of *monitoringpb.Snooze.
type SnoozeIterator struct {
	items    []*monitoringpb.Snooze
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.Snooze, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *SnoozeIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *SnoozeIterator) Next() (*monitoringpb.Snooze, error) {
	var item *monitoringpb.Snooze
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *SnoozeIterator) bufLen() int {
	return len(it.items)
}

func (it *SnoozeIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// TimeSeriesDataIterator manages a stream of *monitoringpb.TimeSeriesData.
type TimeSeriesDataIterator struct {
	items    []*monitoringpb.TimeSeriesData
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.TimeSeriesData, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *TimeSeriesDataIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *TimeSeriesDataIterator) Next() (*monitoringpb.TimeSeriesData, error) {
	var item *monitoringpb.TimeSeriesData
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *TimeSeriesDataIterator) bufLen() int {
	return len(it.items)
}

func (it *TimeSeriesDataIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// TimeSeriesIterator manages a stream of *monitoringpb.TimeSeries.
type TimeSeriesIterator struct {
	items    []*monitoringpb.TimeSeries
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.TimeSeries, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *TimeSeriesIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *TimeSeriesIterator) Next() (*monitoringpb.TimeSeries, error) {
	var item *monitoringpb.TimeSeries
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *TimeSeriesIterator) bufLen() int {
	return len(it.items)
}

func (it *TimeSeriesIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// UptimeCheckConfigIterator manages a stream of *monitoringpb.UptimeCheckConfig.
type UptimeCheckConfigIterator struct {
	items    []*monitoringpb.UptimeCheckConfig
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.UptimeCheckConfig, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *UptimeCheckConfigIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *UptimeCheckConfigIterator) Next() (*monitoringpb.UptimeCheckConfig, error) {
	var item *monitoringpb.UptimeCheckConfig
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *UptimeCheckConfigIterator) bufLen() int {
	return len(it.items)
}

func (it *UptimeCheckConfigIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}

// UptimeCheckIpIterator manages a stream of *monitoringpb.UptimeCheckIp.
type UptimeCheckIpIterator struct {
	items    []*monitoringpb.UptimeCheckIp
	pageInfo *iterator.PageInfo
	nextFunc func() error

	// Response is the raw response for the current page.
	// It must be cast to the RPC response type.
	// Calling Next() or InternalFetch() updates this value.
	Response interface{}

	// InternalFetch is for use by the Google Cloud Libraries only.
	// It is not part of the stable interface of this package.
	//
	// InternalFetch returns results from a single call to the underlying RPC.
	// The number of results is no greater than pageSize.
	// If there are no more results, nextPageToken is empty and err is nil.
	InternalFetch func(pageSize int, pageToken string) (results []*monitoringpb.UptimeCheckIp, nextPageToken string, err error)
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *UptimeCheckIpIterator) PageInfo() *iterator.PageInfo {
	return it.pageInfo
}

// Next returns the next result. Its second return value is iterator.Done if there are no more
// results. Once Next returns Done, all subsequent calls will return Done.
func (it *UptimeCheckIpIterator) Next() (*monitoringpb.UptimeCheckIp, error) {
	var item *monitoringpb.UptimeCheckIp
	if err := it.nextFunc(); err != nil {
		return item, err
	}
	item = it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *UptimeCheckIpIterator) bufLen() int {
	return len(it.items)
}

func (it *UptimeCheckIpIterator) takeBuf() interface{} {
	b := it.items
	it.items = nil
	return b
}
