// Code generated by smithy-go-codegen DO NOT EDIT.

package codecommit

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Renames a repository. The repository name must be unique across the calling AWS
// account. Repository names are limited to 100 alphanumeric, dash, and underscore
// characters, and cannot include certain characters. The suffix .git is
// prohibited. For more information about the limits on repository names, see
// Limits (https://docs.aws.amazon.com/codecommit/latest/userguide/limits.html) in
// the AWS CodeCommit User Guide.
func (c *Client) UpdateRepositoryName(ctx context.Context, params *UpdateRepositoryNameInput, optFns ...func(*Options)) (*UpdateRepositoryNameOutput, error) {
	if params == nil {
		params = &UpdateRepositoryNameInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "UpdateRepositoryName", params, optFns, c.addOperationUpdateRepositoryNameMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*UpdateRepositoryNameOutput)
	out.ResultMetadata = metadata
	return out, nil
}

// Represents the input of an update repository description operation.
type UpdateRepositoryNameInput struct {

	// The new name for the repository.
	//
	// This member is required.
	NewName *string

	// The current name of the repository.
	//
	// This member is required.
	OldName *string

	noSmithyDocumentSerde
}

type UpdateRepositoryNameOutput struct {
	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationUpdateRepositoryNameMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpUpdateRepositoryName{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpUpdateRepositoryName{}, middleware.After)
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
	if err = addOpUpdateRepositoryNameValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opUpdateRepositoryName(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opUpdateRepositoryName(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "codecommit",
		OperationName: "UpdateRepositoryName",
	}
}
