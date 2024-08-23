package client

import (
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	HTTP MetricType = "http"
	SDK  MetricType = "sdk"
	API  MetricType = "api"
)

type Metric interface {
	Type() string
}

type HttpMetric struct {
	TransactionId string  `json:"transactionId"`
	RetryCount    int     `json:"retryCount"`
	Error         error   `json:"error,omitempty"`
	Duration      int64   `json:"duration"`
	ResourcePath  string  `json:"resourcePath"`
	Status        string  `json:"status,omitempty"`
	StatusCode    int     `json:"statusCode,omitempty"`
	HttpRequest   request `json:"request,omitempty"`
}

func (hm *HttpMetric) Type() string {
	return string(HTTP)
}

type ApiMetric struct {
	TransactionId  string         `json:"transactionId"`
	Duration       int64          `json:"duration"`
	ResourcePath   string         `json:"resourcePath"`
	ResultMetadata ResultMetadata `json:"resultMetadata, omitempty"`
	HttpResponse   http.Response  `json:"httpResponse, omitempty"`
}

func (apiMetric *ApiMetric) Type() string {
	return string(API)
}

type SdkMetric struct {
	TransactionId     string     `json:"transactionId"`
	Duration          int64      `json:"duration"`
	ErrorType         string     `json:"errorType"`
	ErrorMessage      string     `json:"errorMessage"`
	ResourcePath      string     `json:"resourcePath"`
	SdkRequestDetails ApiRequest `json:"sdkRequestDetails"`
	SdkResultDetails  ApiResult  `json:"sdkResultDetails"`
}

func (apiMetric *SdkMetric) Type() string {
	return string(SDK)
}

type Process func(metric Metric) interface{}

type MetricType string

var AvailableMetricTypes = []MetricType{HTTP, API, SDK}

type MetricPublisher struct {
	SubscriberMap map[string][]MetricSubscriber
	mux           sync.Mutex
}

type MetricSubscriber struct {
	Process Process
}

func (s *MetricSubscriber) Register(metricType MetricType) {
	metricPublisher.mux.Lock()
	if metricPublisher.SubscriberMap == nil {
		metricPublisher.SubscriberMap = make(map[string][]MetricSubscriber)
	}
	subs := metricPublisher.SubscriberMap[string(metricType)]
	subs = append(subs, *s)
	metricPublisher.SubscriberMap[string(metricType)] = subs
	metricPublisher.mux.Unlock()
}

func (mp *MetricPublisher) publish(metric Metric) {
	for _, sub := range metricPublisher.SubscriberMap[metric.Type()] {
		if sub.Process != nil {
			m := metric //give copy of the object for all subs
			sub.Process(m)
		}
	}
}

func duration(start, end int64) int64 {
	startMillisecond := start / int64(time.Millisecond)
	endMillisecond := end / int64(time.Millisecond)
	return endMillisecond - startMillisecond
}

func generateTransactionId() string {
	return randStringRunes(20)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func buildHttpMetric(transactionId string, resourcePath string, response *http.Response, err error, duration int64, httpRequest request) *HttpMetric {
	retryCount, convErr := strconv.Atoi(response.Header.Get("retryCount"))
	metric := &HttpMetric{
		TransactionId: transactionId,
		RetryCount:    0,
		Error:         err,
		Duration:      duration,
		ResourcePath:  resourcePath,
		HttpRequest:   httpRequest,
		Status:        response.Status,
		StatusCode:    response.StatusCode,
	}
	if convErr == nil {
		metric.RetryCount = retryCount
	}
	return metric
}

func buildSdkMetric(transactionId string, resourcePath string, errorType string, err error, apiRequest ApiRequest, apiResult ApiResult, duration int64) *SdkMetric {
	metric := &SdkMetric{}
	metric.TransactionId = transactionId
	if err != nil {
		metric.ErrorMessage = err.Error()
		metric.ErrorType = errorType
	}
	metric.ResourcePath = resourcePath
	metric.SdkRequestDetails = apiRequest
	metric.Duration = duration
	metric.SdkResultDetails = apiResult
	return metric
}

func buildApiMetric(transactionId string, resourcePath string, duration int64, metadata ResultMetadata, response *http.Response, err error) *ApiMetric {
	metric := &ApiMetric{}
	metric.TransactionId = transactionId
	metric.ResourcePath = resourcePath
	metric.Duration = duration
	if err == nil {
		metric.ResultMetadata = metadata
	} else if ae, ok := err.(*ApiError); ok {
		rm := ResultMetadata{}
		rm.RequestId = ae.RequestId
		rm.ResponseTime = ae.Took
		rm.RateLimitPeriod = response.Header.Get("X-RateLimit-Period-In-Sec")
		rm.RateLimitReason = response.Header.Get("X-RateLimit-Reason")
		rm.RateLimitState = response.Header.Get("X-RateLimit-State")
		rc, convErr := strconv.Atoi(response.Header.Get("retryCount"))
		if convErr == nil {
			rm.RetryCount = rc
		}
		metric.ResultMetadata = rm
	}
	metric.HttpResponse = *response
	return metric
}
