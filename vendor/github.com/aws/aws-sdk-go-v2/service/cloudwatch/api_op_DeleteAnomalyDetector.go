// Code generated by smithy-go-codegen DO NOT EDIT.

package cloudwatch

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Deletes the specified anomaly detection model from your account.
func (c *Client) DeleteAnomalyDetector(ctx context.Context, params *DeleteAnomalyDetectorInput, optFns ...func(*Options)) (*DeleteAnomalyDetectorOutput, error) {
	if params == nil {
		params = &DeleteAnomalyDetectorInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DeleteAnomalyDetector", params, optFns, c.addOperationDeleteAnomalyDetectorMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DeleteAnomalyDetectorOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DeleteAnomalyDetectorInput struct {

	// The metric dimensions associated with the anomaly detection model to delete.
	//
	// Deprecated: Use SingleMetricAnomalyDetector.
	Dimensions []types.Dimension

	// The metric math anomaly detector to be deleted. When using
	// MetricMathAnomalyDetector, you cannot include following parameters in the same
	// operation:
	//
	// * Dimensions,
	//
	// * MetricName
	//
	// * Namespace
	//
	// * Stat
	//
	// * the
	// SingleMetricAnomalyDetector parameters of DeleteAnomalyDetectorInput
	//
	// Instead,
	// specify the metric math anomaly detector attributes as part of the
	// MetricMathAnomalyDetector property.
	MetricMathAnomalyDetector *types.MetricMathAnomalyDetector

	// The metric name associated with the anomaly detection model to delete.
	//
	// Deprecated: Use SingleMetricAnomalyDetector.
	MetricName *string

	// The namespace associated with the anomaly detection model to delete.
	//
	// Deprecated: Use SingleMetricAnomalyDetector.
	Namespace *string

	// A single metric anomaly detector to be deleted. When using
	// SingleMetricAnomalyDetector, you cannot include the following parameters in the
	// same operation:
	//
	// * Dimensions,
	//
	// * MetricName
	//
	// * Namespace
	//
	// * Stat
	//
	// * the
	// MetricMathAnomalyDetector parameters of DeleteAnomalyDetectorInput
	//
	// Instead,
	// specify the single metric anomaly detector attributes as part of the
	// SingleMetricAnomalyDetector property.
	SingleMetricAnomalyDetector *types.SingleMetricAnomalyDetector

	// The statistic associated with the anomaly detection model to delete.
	//
	// Deprecated: Use SingleMetricAnomalyDetector.
	Stat *string

	noSmithyDocumentSerde
}

type DeleteAnomalyDetectorOutput struct {
	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationDeleteAnomalyDetectorMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsquery_serializeOpDeleteAnomalyDetector{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsquery_deserializeOpDeleteAnomalyDetector{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpDeleteAnomalyDetectorValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDeleteAnomalyDetector(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opDeleteAnomalyDetector(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "monitoring",
		OperationName: "DeleteAnomalyDetector",
	}
}
