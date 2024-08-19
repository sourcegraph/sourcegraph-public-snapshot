// Code generated by smithy-go-codegen DO NOT EDIT.

package s3

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	s3cust "github.com/aws/aws-sdk-go-v2/service/s3/internal/customizations"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Removes the entire tag set from the specified object. For more information about
// managing object tags, see  Object Tagging
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/object-tagging.html). To use
// this operation, you must have permission to perform the s3:DeleteObjectTagging
// action. To delete tags of a specific object version, add the versionId query
// parameter in the request. You will need permission for the
// s3:DeleteObjectVersionTagging action. The following operations are related to
// DeleteBucketMetricsConfiguration:
//
// * PutObjectTagging
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObjectTagging.html)
//
// *
// GetObjectTagging
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObjectTagging.html)
func (c *Client) DeleteObjectTagging(ctx context.Context, params *DeleteObjectTaggingInput, optFns ...func(*Options)) (*DeleteObjectTaggingOutput, error) {
	if params == nil {
		params = &DeleteObjectTaggingInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DeleteObjectTagging", params, optFns, c.addOperationDeleteObjectTaggingMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DeleteObjectTaggingOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DeleteObjectTaggingInput struct {

	// The bucket name containing the objects from which to remove the tags. When using
	// this action with an access point, you must direct requests to the access point
	// hostname. The access point hostname takes the form
	// AccessPointName-AccountId.s3-accesspoint.Region.amazonaws.com. When using this
	// action with an access point through the Amazon Web Services SDKs, you provide
	// the access point ARN in place of the bucket name. For more information about
	// access point ARNs, see Using access points
	// (https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-access-points.html)
	// in the Amazon S3 User Guide. When using this action with Amazon S3 on Outposts,
	// you must direct requests to the S3 on Outposts hostname. The S3 on Outposts
	// hostname takes the form
	// AccessPointName-AccountId.outpostID.s3-outposts.Region.amazonaws.com. When using
	// this action with S3 on Outposts through the Amazon Web Services SDKs, you
	// provide the Outposts bucket ARN in place of the bucket name. For more
	// information about S3 on Outposts ARNs, see Using Amazon S3 on Outposts
	// (https://docs.aws.amazon.com/AmazonS3/latest/userguide/S3onOutposts.html) in the
	// Amazon S3 User Guide.
	//
	// This member is required.
	Bucket *string

	// The key that identifies the object in the bucket from which to remove all tags.
	//
	// This member is required.
	Key *string

	// The account ID of the expected bucket owner. If the bucket is owned by a
	// different account, the request fails with the HTTP status code 403 Forbidden
	// (access denied).
	ExpectedBucketOwner *string

	// The versionId of the object that the tag-set will be removed from.
	VersionId *string

	noSmithyDocumentSerde
}

type DeleteObjectTaggingOutput struct {

	// The versionId of the object the tag-set was removed from.
	VersionId *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationDeleteObjectTaggingMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestxml_serializeOpDeleteObjectTagging{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestxml_deserializeOpDeleteObjectTagging{}, middleware.After)
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
	if err = swapWithCustomHTTPSignerMiddleware(stack, options); err != nil {
		return err
	}
	if err = addOpDeleteObjectTaggingValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDeleteObjectTagging(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addMetadataRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addDeleteObjectTaggingUpdateEndpoint(stack, options); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = v4.AddContentSHA256HeaderMiddleware(stack); err != nil {
		return err
	}
	if err = disableAcceptEncodingGzip(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opDeleteObjectTagging(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "s3",
		OperationName: "DeleteObjectTagging",
	}
}

// getDeleteObjectTaggingBucketMember returns a pointer to string denoting a
// provided bucket member valueand a boolean indicating if the input has a modeled
// bucket name,
func getDeleteObjectTaggingBucketMember(input interface{}) (*string, bool) {
	in := input.(*DeleteObjectTaggingInput)
	if in.Bucket == nil {
		return nil, false
	}
	return in.Bucket, true
}
func addDeleteObjectTaggingUpdateEndpoint(stack *middleware.Stack, options Options) error {
	return s3cust.UpdateEndpoint(stack, s3cust.UpdateEndpointOptions{
		Accessor: s3cust.UpdateEndpointParameterAccessor{
			GetBucketFromInput: getDeleteObjectTaggingBucketMember,
		},
		UsePathStyle:                   options.UsePathStyle,
		UseAccelerate:                  options.UseAccelerate,
		SupportsAccelerate:             true,
		TargetS3ObjectLambda:           false,
		EndpointResolver:               options.EndpointResolver,
		EndpointResolverOptions:        options.EndpointOptions,
		UseARNRegion:                   options.UseARNRegion,
		DisableMultiRegionAccessPoints: options.DisableMultiRegionAccessPoints,
	})
}
