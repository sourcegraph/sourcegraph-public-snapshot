package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type OpsGenieClient struct {
	RetryableClient *retryablehttp.Client
	Config          *Config
}

type request struct {
	*retryablehttp.Request
}

type ApiRequest interface {
	Validate() error
	ResourcePath() string
	Method() string
	Metadata(apiRequest ApiRequest) map[string]interface{}
	RequestParams() map[string]string
}

var metricPublisher = &MetricPublisher{}

type BaseRequest struct {
}

func (r *BaseRequest) Metadata(apiRequest ApiRequest) map[string]interface{} {
	headers := make(map[string]interface{})
	if apiRequest.Method() != http.MethodGet && apiRequest.Method() != http.MethodDelete {
		headers["Content-Type"] = "application/json; charset=utf-8"
	} else if apiRequest.Method() == http.MethodGet {
		headers["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"
	}
	return headers
}

func (r *BaseRequest) RequestParams() map[string]string {
	return nil
}

type AsyncApiResult interface {
	ApiResult
	RetrieveStatus(ctx context.Context, request ApiRequest, result ApiResult) error
}

type AsyncBaseResult struct {
	Client *OpsGenieClient
}

func (ar *AsyncBaseResult) RetrieveStatus(ctx context.Context, request ApiRequest, result ApiResult) error {

	if ctx == nil {
		ctx = context.Background()
	}

	for i := 0; ; i++ {

		err := ar.Client.Exec(ctx, request, result)
		if err != nil {
			apiErr, ok := err.(*ApiError)
			if !ok ||
				apiErr.StatusCode != 404 ||
				apiErr.ErrorHeader != "RequestNotProcessed" ||
				i >= ar.Client.RetryableClient.RetryMax {
				return err
			}

		} else {
			return nil
		}

		wait := retryablehttp.DefaultBackoff(
			ar.Client.RetryableClient.RetryWaitMin,
			ar.Client.RetryableClient.RetryWaitMax,
			i, nil)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

type ApiResult interface {
	Parse(response *http.Response, result ApiResult) error
	ValidateResultMetadata() error
	setResultMetadata(metadata *ResultMetadata) *ResultMetadata
}

type ResultMetadata struct {
	RequestId       string  `json:"requestId"`
	ResponseTime    float32 `json:"took"`
	RateLimitState  string
	RateLimitReason string
	RateLimitPeriod string
	RetryCount      int
}

func (rm *ResultMetadata) setResultMetadata(metadata *ResultMetadata) *ResultMetadata {
	if len(metadata.RequestId) > 0 {
		rm.RequestId = metadata.RequestId
	}
	if metadata.ResponseTime != 0 {
		rm.ResponseTime = metadata.ResponseTime
	}
	rm.RateLimitState = metadata.RateLimitState
	rm.RateLimitReason = metadata.RateLimitReason
	rm.RateLimitPeriod = metadata.RateLimitPeriod
	rm.RetryCount = metadata.RetryCount
	return rm
}

func (rm *ResultMetadata) ValidateResultMetadata() error {
	unsetFields := ""

	if len(rm.RequestId) == 0 {
		unsetFields = " requestId,"
	}
	if len(rm.RateLimitState) == 0 {
		unsetFields = unsetFields + " rate limit state,"
	}
	if rm.ResponseTime == 0 {
		unsetFields = unsetFields + " response time,"
	}

	if unsetFields != "" {
		unsetFields = unsetFields[:len(unsetFields)-1] + "."
		return errors.New("Could not set" + unsetFields)
	}

	return nil
}

var UserAgentHeader string

const Version = "2.0.0"

func setConfiguration(opsGenieClient *OpsGenieClient, cfg *Config) {
	opsGenieClient.RetryableClient.ErrorHandler = opsGenieClient.defineErrorHandler
	if cfg.OpsGenieAPIURL == "" {
		cfg.OpsGenieAPIURL = API_URL
	}
	if cfg.HttpClient != nil {
		opsGenieClient.RetryableClient.HTTPClient = cfg.HttpClient
	}
	if cfg.RequestTimeout != 0 {
		opsGenieClient.RetryableClient.HTTPClient.Timeout = cfg.RequestTimeout
	}
	if cfg.ProxyConfiguration != nil {
		setProxySettings(opsGenieClient)
	}
	opsGenieClient.Config.apiUrl = string(cfg.OpsGenieAPIURL)
}

func setLogger(conf *Config) {
	// if user has already set logger, skip
	if conf.Logger != nil {
		return
	}

	// otherwise, create a new logger for the user
	logger := logrus.New()
	// set log level if user has specified one
	if conf.LogLevel != (logrus.Level(0)) {
		logger.SetLevel(conf.LogLevel)
	}
	logger.SetFormatter(
		&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339Nano,
		},
	)
	conf.Logger = logger

}

func setRetryPolicy(opsGenieClient *OpsGenieClient, cfg *Config) {
	//custom backoff
	if cfg.Backoff != nil {
		opsGenieClient.RetryableClient.Backoff = cfg.Backoff
	}

	//custom retry policy
	if cfg.RetryPolicy != nil {
		opsGenieClient.RetryableClient.CheckRetry = cfg.RetryPolicy
	} else {
		opsGenieClient.RetryableClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (b bool, e error) {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}

			if err != nil {
				return true, err
			}
			// Check the response code. We retry on 500-range responses to allow
			// the server time to recover, as 500's are typically not permanent
			// errors and may relate to outages on the server side. This will catch
			// invalid response codes as well, like 0 and 999.
			if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != 501) {
				return true, nil
			}
			if resp.StatusCode == 429 {
				return true, nil
			}

			return false, nil
		}
	}

	if cfg.RetryCount != 0 {
		opsGenieClient.RetryableClient.RetryMax = cfg.RetryCount
	} else {
		opsGenieClient.RetryableClient.RetryMax = 4
	}
}

func NewOpsGenieClient(cfg *Config) (*OpsGenieClient, error) {
	UserAgentHeader = fmt.Sprintf("opsgenie-go-sdk-%s %s (%s/%s)", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	opsGenieClient := &OpsGenieClient{
		Config:          cfg,
		RetryableClient: retryablehttp.NewClient(),
	}
	if cfg.Validate() != nil {
		return nil, cfg.Validate()
	}
	setConfiguration(opsGenieClient, cfg)
	opsGenieClient.RetryableClient.Logger = nil //disable retryableClient's uncustomizable logging
	setLogger(cfg)
	setRetryPolicy(opsGenieClient, cfg)
	printInfoLog(opsGenieClient)
	return opsGenieClient, nil
}

func printInfoLog(client *OpsGenieClient) {
	client.Config.Logger.Infof("Client is configured with ApiUrl: %s, RetryMaxCount: %v",
		client.Config.OpsGenieAPIURL,
		client.RetryableClient.RetryMax)
}

func (cli *OpsGenieClient) defineErrorHandler(resp *http.Response, err error, numTries int) (*http.Response, error) {
	if err != nil {
		cli.Config.Logger.Errorf("Unable to send the request %s ", err.Error())
		if err == context.DeadlineExceeded {
			return nil, err
		}
		return nil, err
	}
	resp.Header.Add("retryCount", strconv.Itoa(numTries))
	cli.Config.Logger.Errorf("Failed to process request after %d attempts.", numTries)
	return resp, nil
}

func (cli *OpsGenieClient) do(request *request) (*http.Response, error) {
	return cli.RetryableClient.Do(request.Request)
}

func setResultMetadata(httpResponse *http.Response, result ApiResult) *ResultMetadata {
	responseTime := httpResponse.Header.Get("X-Response-Time")

	retryCount, err := strconv.Atoi(httpResponse.Header.Get("retryCount"))
	responseTimeInFloat, err2 := strconv.ParseFloat(responseTime, 32)
	resultMetadata := &ResultMetadata{
		RequestId:       httpResponse.Header.Get("X-Request-Id"),
		RateLimitState:  httpResponse.Header.Get("X-RateLimit-State"),
		RateLimitReason: httpResponse.Header.Get("X-RateLimit-Reason"),
		RateLimitPeriod: httpResponse.Header.Get("X-RateLimit-Period-In-Sec"),
	}
	if err == nil {
		resultMetadata.RetryCount = retryCount
	}
	if err2 == nil {
		resultMetadata.ResponseTime = float32(responseTimeInFloat)
	}
	return result.setResultMetadata(resultMetadata)
}

type ApiError struct {
	error
	Message     string            `json:"message"`
	Took        float32           `json:"took"`
	RequestId   string            `json:"requestId"`
	Errors      map[string]string `json:"errors"`
	StatusCode  int
	ErrorHeader string
}

func (ar *ApiError) Error() string {
	errMessage := "Error occurred with Status code: " + strconv.Itoa(ar.StatusCode) + ", " +
		"Message: " + ar.Message + ", " +
		"Took: " + fmt.Sprintf("%f", ar.Took) + ", " +
		"RequestId: " + ar.RequestId
	if ar.ErrorHeader != "" {
		errMessage = errMessage + ", Error Header: " + ar.ErrorHeader
	}
	if ar.Errors != nil {
		errMessage = errMessage + ", Error Detail: " + fmt.Sprintf("%v", ar.Errors)
	}
	return errMessage
}

func handleErrorIfExist(response *http.Response) error {
	if response != nil && response.StatusCode >= 400 {
		apiError := &ApiError{}
		apiError.StatusCode = response.StatusCode
		apiError.ErrorHeader = response.Header.Get("X-Opsgenie-Errortype")
		body, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(body, apiError)
		return apiError
	}
	return nil
}

func (cli *OpsGenieClient) buildHttpRequest(apiRequest ApiRequest) (*request, error) {
	var buf io.ReadWriter
	var contentType = new(string)
	var err error
	var req = new(retryablehttp.Request)

	details := apiRequest.Metadata(apiRequest)
	if values, ok := details["form-data-values"].(map[string]io.Reader); ok {
		setBodyAsFormData(&buf, values, contentType)
	} else if apiRequest.Method() != http.MethodGet && apiRequest.Method() != http.MethodDelete {
		err = setBodyAsJson(&buf, apiRequest, contentType, details)
	}
	if err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	for key, value := range apiRequest.RequestParams() {
		queryParams.Add(key, value)
	}

	req, err = retryablehttp.NewRequest(apiRequest.Method(), buildRequestUrl(cli, apiRequest, queryParams), buf)
	if err != nil {
		return nil, err
	}

	if contentType != nil {
		req.Header.Add("Content-Type", *contentType)
	} else {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "GenieKey "+cli.Config.ApiKey)
	req.Header.Add("User-Agent", UserAgentHeader)

	return &request{req}, err

}

func buildRequestUrl(cli *OpsGenieClient, apiRequest ApiRequest, queryParams url.Values) string {
	requestUrl := url.URL{
		Scheme:   string(Https),
		Host:     cli.Config.apiUrl,
		Path:     apiRequest.ResourcePath(),
		RawQuery: queryParams.Encode(),
	}
	//test purposes only
	if !strings.Contains(cli.Config.apiUrl, "api") {
		requestUrl.Scheme = "http"
	}
	//
	return requestUrl.String()
}

func setProxySettings(cli *OpsGenieClient) {
	proxy := cli.Config.ProxyConfiguration.Host
	if cli.Config.ProxyConfiguration.Port != 0 {
		proxy = proxy + ":" + strconv.Itoa(cli.Config.ProxyConfiguration.Port)
	}
	proxyUrl := &url.URL{
		Host:   proxy,
		Scheme: string(cli.Config.ProxyConfiguration.Protocol),
	}
	if cli.Config.ProxyConfiguration.Username != "" {
		proxyUrl.User = url.UserPassword(cli.Config.ProxyConfiguration.Username, cli.Config.ProxyConfiguration.Password)
	}
	cli.RetryableClient.HTTPClient.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
}

func setBodyAsJson(buf *io.ReadWriter, apiRequest ApiRequest, contentType *string, details map[string]interface{}) error {
	*buf = new(bytes.Buffer)
	*contentType = details["Content-Type"].(string)

	err := json.NewEncoder(*buf).Encode(apiRequest)
	if err != nil {
		return err
	}

	return nil
}

func setBodyAsFormData(buf *io.ReadWriter, values map[string]io.Reader, contentType *string) error {

	*buf = new(bytes.Buffer)
	writer := multipart.NewWriter(*buf)
	defer writer.Close()

	for key, reader := range values {
		var part io.Writer
		var err error
		if file, ok := reader.(*os.File); ok {
			fileStat, err := file.Stat()
			if err != nil {
				return err
			}
			part, err = writer.CreateFormFile(key, fileStat.Name())
			if err != nil {
				return err
			}
		} else {
			part, err = writer.CreateFormField(key)
			if err != nil {
				return err
			}
		}
		io.Copy(part, reader)
	}

	*contentType = writer.FormDataContentType()
	return nil
}

func (cli *OpsGenieClient) Exec(ctx context.Context, request ApiRequest, result ApiResult) error {
	startTime := time.Now().UnixNano()
	transactionId := generateTransactionId()
	cli.Config.Logger.Debugf("Starting to process Request %+v: to send: %s", request, request.ResourcePath())
	if err := request.Validate(); err != nil {
		cli.Config.Logger.Errorf("Request validation err: %s ", err.Error())
		metricPublisher.publish(buildSdkMetric(transactionId, request.ResourcePath(), "request-validation-error", err, request, result, duration(startTime, time.Now().UnixNano())))
		return err
	}
	req, err := cli.buildHttpRequest(request)
	if err != nil {
		cli.Config.Logger.Errorf("Could not create request: %s", err.Error())
		metricPublisher.publish(buildSdkMetric(transactionId, request.ResourcePath(), "sdk-error", err, request, result, duration(startTime, time.Now().UnixNano())))
		return err
	}
	if ctx != nil {
		req.WithContext(ctx)
	}

	response, err := cli.do(req)
	if response != nil {
		metricPublisher.publish(buildHttpMetric(transactionId, request.ResourcePath(), response, err, duration(startTime, time.Now().UnixNano()), *req))
	}
	if err != nil {
		cli.Config.Logger.Errorf(err.Error())
		return err
	}

	defer response.Body.Close()

	err = handleErrorIfExist(response)
	if err != nil {
		cli.Config.Logger.Errorf(err.Error())
		metricPublisher.publish(buildApiMetric(transactionId, request.ResourcePath(), duration(startTime, time.Now().UnixNano()), *setResultMetadata(response, result), response, err))
		metricPublisher.publish(buildSdkMetric(transactionId, request.ResourcePath(), "api-error", err, request, result, duration(startTime, time.Now().UnixNano())))
		return err
	}

	err = result.Parse(response, result)
	if err != nil {
		cli.Config.Logger.Errorf(err.Error())
		metricPublisher.publish(buildSdkMetric(transactionId, request.ResourcePath(), "http-response-parsing-error", err, request, result, duration(startTime, time.Now().UnixNano())))
		return err
	}

	rm := setResultMetadata(response, result)
	metricPublisher.publish(buildApiMetric(transactionId, request.ResourcePath(), duration(startTime, time.Now().UnixNano()), *rm, response, nil))
	err = result.ValidateResultMetadata()
	if err != nil {
		cli.Config.Logger.Warn(err.Error())
	}
	metricPublisher.publish(buildSdkMetric(transactionId, request.ResourcePath(), "", nil, request, result, duration(startTime, time.Now().UnixNano())))
	cli.Config.Logger.Debugf("Request processed. The result: %+v", result)
	return nil
}

func shouldDataIgnored(result ApiResult) bool {
	resultType := reflect.TypeOf(result)
	elem := resultType.Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if strings.Contains(field.Tag.Get("json"), "data") {
			return false
		}
	}
	return true
}

func (rm *ResultMetadata) Parse(response *http.Response, result ApiResult) error {

	var payload []byte
	if response == nil {
		return errors.New("No response received")
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	payload = body

	if shouldDataIgnored(result) {
		resultMap := make(map[string]interface{})
		err = json.Unmarshal(body, &resultMap)
		if err != nil {
			return handleParsingErrors(err)
		}
		if value, ok := resultMap["data"]; ok {
			payload, err = json.Marshal(value)
			if err != nil {
				return handleParsingErrors(err)
			}
		}
	}

	err = json.Unmarshal(payload, result)

	if err != nil {
		return handleParsingErrors(err)
	}

	return nil
}

func handleParsingErrors(err error) error {
	message := "Response could not be parsed, " + err.Error()
	return errors.New(message)
}
