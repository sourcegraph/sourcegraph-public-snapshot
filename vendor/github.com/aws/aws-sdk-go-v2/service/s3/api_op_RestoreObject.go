// Code generated by smithy-go-codegen DO NOT EDIT.

package s3

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	internalChecksum "github.com/aws/aws-sdk-go-v2/service/internal/checksum"
	s3cust "github.com/aws/aws-sdk-go-v2/service/s3/internal/customizations"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Restores an archived copy of an object back into Amazon S3 This action is not
// supported by Amazon S3 on Outposts. This action performs the following types of
// requests:
//
// * select - Perform a select query on an archived object
//
// * restore an
// archive - Restore an archived object
//
// To use this operation, you must have
// permissions to perform the s3:RestoreObject action. The bucket owner has this
// permission by default and can grant this permission to others. For more
// information about permissions, see Permissions Related to Bucket Subresource
// Operations
// (https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-with-s3-actions.html#using-with-s3-actions-related-to-bucket-subresources)
// and Managing Access Permissions to Your Amazon S3 Resources
// (https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-access-control.html)
// in the Amazon S3 User Guide. Querying Archives with Select Requests You use a
// select type of request to perform SQL queries on archived objects. The archived
// objects that are being queried by the select request must be formatted as
// uncompressed comma-separated values (CSV) files. You can run queries and custom
// analytics on your archived data without having to restore your data to a hotter
// Amazon S3 tier. For an overview about select requests, see Querying Archived
// Objects
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/querying-glacier-archives.html)
// in the Amazon S3 User Guide. When making a select request, do the following:
//
// *
// Define an output location for the select query's output. This must be an Amazon
// S3 bucket in the same Amazon Web Services Region as the bucket that contains the
// archive object that is being queried. The Amazon Web Services account that
// initiates the job must have permissions to write to the S3 bucket. You can
// specify the storage class and encryption for the output objects stored in the
// bucket. For more information about output, see Querying Archived Objects
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/querying-glacier-archives.html)
// in the Amazon S3 User Guide. For more information about the S3 structure in the
// request body, see the following:
//
// * PutObject
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html)
//
// * Managing
// Access with ACLs
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/S3_ACLs_UsingACLs.html) in the
// Amazon S3 User Guide
//
// * Protecting Data Using Server-Side Encryption
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/serv-side-encryption.html) in
// the Amazon S3 User Guide
//
// * Define the SQL expression for the SELECT type of
// restoration for your query in the request body's SelectParameters structure. You
// can use expressions like the following examples.
//
// * The following expression
// returns all records from the specified object. SELECT * FROM Object
//
// * Assuming
// that you are not using any headers for data stored in the object, you can
// specify columns with positional headers. SELECT s._1, s._2 FROM Object s WHERE
// s._3 > 100
//
// * If you have headers and you set the fileHeaderInfo in the CSV
// structure in the request body to USE, you can specify headers in the query. (If
// you set the fileHeaderInfo field to IGNORE, the first row is skipped for the
// query.) You cannot mix ordinal positions with header column names. SELECT s.Id,
// s.FirstName, s.SSN FROM S3Object s
//
// For more information about using SQL with S3
// Glacier Select restore, see SQL Reference for Amazon S3 Select and S3 Glacier
// Select
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-glacier-select-sql-reference.html)
// in the Amazon S3 User Guide. When making a select request, you can also do the
// following:
//
// * To expedite your queries, specify the Expedited tier. For more
// information about tiers, see "Restoring Archives," later in this topic.
//
// *
// Specify details about the data serialization format of both the input object
// that is being queried and the serialization of the CSV-encoded query
// results.
//
// The following are additional important facts about the select
// feature:
//
// * The output results are new Amazon S3 objects. Unlike archive
// retrievals, they are stored until explicitly deleted-manually or through a
// lifecycle policy.
//
// * You can issue more than one select request on the same
// Amazon S3 object. Amazon S3 doesn't deduplicate requests, so avoid issuing
// duplicate requests.
//
// * Amazon S3 accepts a select request even if the object has
// already been restored. A select request doesn’t return error response
// 409.
//
// Restoring objects Objects that you archive to the S3 Glacier or S3 Glacier
// Deep Archive storage class, and S3 Intelligent-Tiering Archive or S3
// Intelligent-Tiering Deep Archive tiers are not accessible in real time. For
// objects in Archive Access or Deep Archive Access tiers you must first initiate a
// restore request, and then wait until the object is moved into the Frequent
// Access tier. For objects in S3 Glacier or S3 Glacier Deep Archive storage
// classes you must first initiate a restore request, and then wait until a
// temporary copy of the object is available. To access an archived object, you
// must restore the object for the duration (number of days) that you specify. To
// restore a specific object version, you can provide a version ID. If you don't
// provide a version ID, Amazon S3 restores the current version. When restoring an
// archived object (or using a select request), you can specify one of the
// following data access tier options in the Tier element of the request body:
//
// *
// Expedited - Expedited retrievals allow you to quickly access your data stored in
// the S3 Glacier storage class or S3 Intelligent-Tiering Archive tier when
// occasional urgent requests for a subset of archives are required. For all but
// the largest archived objects (250 MB+), data accessed using Expedited retrievals
// is typically made available within 1–5 minutes. Provisioned capacity ensures
// that retrieval capacity for Expedited retrievals is available when you need it.
// Expedited retrievals and provisioned capacity are not available for objects
// stored in the S3 Glacier Deep Archive storage class or S3 Intelligent-Tiering
// Deep Archive tier.
//
// * Standard - Standard retrievals allow you to access any of
// your archived objects within several hours. This is the default option for
// retrieval requests that do not specify the retrieval option. Standard retrievals
// typically finish within 3–5 hours for objects stored in the S3 Glacier storage
// class or S3 Intelligent-Tiering Archive tier. They typically finish within 12
// hours for objects stored in the S3 Glacier Deep Archive storage class or S3
// Intelligent-Tiering Deep Archive tier. Standard retrievals are free for objects
// stored in S3 Intelligent-Tiering.
//
// * Bulk - Bulk retrievals are the lowest-cost
// retrieval option in S3 Glacier, enabling you to retrieve large amounts, even
// petabytes, of data inexpensively. Bulk retrievals typically finish within 5–12
// hours for objects stored in the S3 Glacier storage class or S3
// Intelligent-Tiering Archive tier. They typically finish within 48 hours for
// objects stored in the S3 Glacier Deep Archive storage class or S3
// Intelligent-Tiering Deep Archive tier. Bulk retrievals are free for objects
// stored in S3 Intelligent-Tiering.
//
// For more information about archive retrieval
// options and provisioned capacity for Expedited data access, see Restoring
// Archived Objects
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/restoring-objects.html) in the
// Amazon S3 User Guide. You can use Amazon S3 restore speed upgrade to change the
// restore speed to a faster speed while it is in progress. For more information,
// see  Upgrading the speed of an in-progress restore
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/restoring-objects.html#restoring-objects-upgrade-tier.title.html)
// in the Amazon S3 User Guide. To get the status of object restoration, you can
// send a HEAD request. Operations return the x-amz-restore header, which provides
// information about the restoration status, in the response. You can use Amazon S3
// event notifications to notify you when a restore is initiated or completed. For
// more information, see Configuring Amazon S3 Event Notifications
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html) in the
// Amazon S3 User Guide. After restoring an archived object, you can update the
// restoration period by reissuing the request with a new period. Amazon S3 updates
// the restoration period relative to the current time and charges only for the
// request-there are no data transfer charges. You cannot update the restoration
// period when Amazon S3 is actively processing your current restore request for
// the object. If your bucket has a lifecycle configuration with a rule that
// includes an expiration action, the object expiration overrides the life span
// that you specify in a restore request. For example, if you restore an object
// copy for 10 days, but the object is scheduled to expire in 3 days, Amazon S3
// deletes the object in 3 days. For more information about lifecycle
// configuration, see PutBucketLifecycleConfiguration
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLifecycleConfiguration.html)
// and Object Lifecycle Management
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html) in
// Amazon S3 User Guide. Responses A successful action returns either the 200 OK or
// 202 Accepted status code.
//
// * If the object is not previously restored, then
// Amazon S3 returns 202 Accepted in the response.
//
// * If the object is previously
// restored, Amazon S3 returns 200 OK in the response.
//
// # Special Errors
//
// * Code:
// RestoreAlreadyInProgress
//
// * Cause: Object restore is already in progress. (This
// error does not apply to SELECT type requests.)
//
// * HTTP Status Code: 409
// Conflict
//
// * SOAP Fault Code Prefix: Client
//
// * Code:
// GlacierExpeditedRetrievalNotAvailable
//
// * Cause: expedited retrievals are
// currently not available. Try again later. (Returned if there is insufficient
// capacity to process the Expedited request. This error applies only to Expedited
// retrievals and not to S3 Standard or Bulk retrievals.)
//
// * HTTP Status Code:
// 503
//
// * SOAP Fault Code Prefix: N/A
//
// # Related Resources
//
// *
// PutBucketLifecycleConfiguration
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLifecycleConfiguration.html)
//
// *
// GetBucketNotificationConfiguration
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketNotificationConfiguration.html)
//
// *
// SQL Reference for Amazon S3 Select and S3 Glacier Select
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-glacier-select-sql-reference.html)
// in the Amazon S3 User Guide
func (c *Client) RestoreObject(ctx context.Context, params *RestoreObjectInput, optFns ...func(*Options)) (*RestoreObjectOutput, error) {
	if params == nil {
		params = &RestoreObjectInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "RestoreObject", params, optFns, c.addOperationRestoreObjectMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*RestoreObjectOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type RestoreObjectInput struct {

	// The bucket name containing the object to restore. When using this action with an
	// access point, you must direct requests to the access point hostname. The access
	// point hostname takes the form
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

	// Object key for which the action was initiated.
	//
	// This member is required.
	Key *string

	// Indicates the algorithm used to create the checksum for the object when using
	// the SDK. This header will not provide any additional functionality if not using
	// the SDK. When sending this header, there must be a corresponding x-amz-checksum
	// or x-amz-trailer header sent. Otherwise, Amazon S3 fails the request with the
	// HTTP status code 400 Bad Request. For more information, see Checking object
	// integrity
	// (https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html)
	// in the Amazon S3 User Guide. If you provide an individual checksum, Amazon S3
	// ignores any provided ChecksumAlgorithm parameter.
	ChecksumAlgorithm types.ChecksumAlgorithm

	// The account ID of the expected bucket owner. If the bucket is owned by a
	// different account, the request fails with the HTTP status code 403 Forbidden
	// (access denied).
	ExpectedBucketOwner *string

	// Confirms that the requester knows that they will be charged for the request.
	// Bucket owners need not specify this parameter in their requests. For information
	// about downloading objects from Requester Pays buckets, see Downloading Objects
	// in Requester Pays Buckets
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/ObjectsinRequesterPaysBuckets.html)
	// in the Amazon S3 User Guide.
	RequestPayer types.RequestPayer

	// Container for restore job parameters.
	RestoreRequest *types.RestoreRequest

	// VersionId used to reference a specific version of the object.
	VersionId *string

	noSmithyDocumentSerde
}

type RestoreObjectOutput struct {

	// If present, indicates that the requester was successfully charged for the
	// request.
	RequestCharged types.RequestCharged

	// Indicates the path in the provided S3 output location where Select results will
	// be restored to.
	RestoreOutputPath *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationRestoreObjectMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestxml_serializeOpRestoreObject{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestxml_deserializeOpRestoreObject{}, middleware.After)
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
	if err = addOpRestoreObjectValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opRestoreObject(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addMetadataRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addRestoreObjectInputChecksumMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addRestoreObjectUpdateEndpoint(stack, options); err != nil {
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

func newServiceMetadataMiddleware_opRestoreObject(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "s3",
		OperationName: "RestoreObject",
	}
}

// getRestoreObjectRequestAlgorithmMember gets the request checksum algorithm value
// provided as input.
func getRestoreObjectRequestAlgorithmMember(input interface{}) (string, bool) {
	in := input.(*RestoreObjectInput)
	if len(in.ChecksumAlgorithm) == 0 {
		return "", false
	}
	return string(in.ChecksumAlgorithm), true
}

func addRestoreObjectInputChecksumMiddlewares(stack *middleware.Stack, options Options) error {
	return internalChecksum.AddInputMiddleware(stack, internalChecksum.InputMiddlewareOptions{
		GetAlgorithm:                     getRestoreObjectRequestAlgorithmMember,
		RequireChecksum:                  false,
		EnableTrailingChecksum:           false,
		EnableComputeSHA256PayloadHash:   true,
		EnableDecodedContentLengthHeader: true,
	})
}

// getRestoreObjectBucketMember returns a pointer to string denoting a provided
// bucket member valueand a boolean indicating if the input has a modeled bucket
// name,
func getRestoreObjectBucketMember(input interface{}) (*string, bool) {
	in := input.(*RestoreObjectInput)
	if in.Bucket == nil {
		return nil, false
	}
	return in.Bucket, true
}
func addRestoreObjectUpdateEndpoint(stack *middleware.Stack, options Options) error {
	return s3cust.UpdateEndpoint(stack, s3cust.UpdateEndpointOptions{
		Accessor: s3cust.UpdateEndpointParameterAccessor{
			GetBucketFromInput: getRestoreObjectBucketMember,
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
