// Code generated by smithy-go-codegen DO NOT EDIT.

package kms

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Returns a random byte string that is cryptographically secure. By default, the
// random byte string is generated in KMS. To generate the byte string in the
// CloudHSM cluster that is associated with a custom key store
// (https://docs.aws.amazon.com/kms/latest/developerguide/custom-key-store-overview.html),
// specify the custom key store ID. Applications in Amazon Web Services Nitro
// Enclaves can call this operation by using the Amazon Web Services Nitro Enclaves
// Development Kit (https://github.com/aws/aws-nitro-enclaves-sdk-c). For
// information about the supporting parameters, see How Amazon Web Services Nitro
// Enclaves use KMS
// (https://docs.aws.amazon.com/kms/latest/developerguide/services-nitro-enclaves.html)
// in the Key Management Service Developer Guide. For more information about
// entropy and random number generation, see Key Management Service Cryptographic
// Details (https://docs.aws.amazon.com/kms/latest/cryptographic-details/).
// Required permissions: kms:GenerateRandom
// (https://docs.aws.amazon.com/kms/latest/developerguide/kms-api-permissions-reference.html)
// (IAM policy)
func (c *Client) GenerateRandom(ctx context.Context, params *GenerateRandomInput, optFns ...func(*Options)) (*GenerateRandomOutput, error) {
	if params == nil {
		params = &GenerateRandomInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GenerateRandom", params, optFns, c.addOperationGenerateRandomMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GenerateRandomOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GenerateRandomInput struct {

	// Generates the random byte string in the CloudHSM cluster that is associated with
	// the specified custom key store
	// (https://docs.aws.amazon.com/kms/latest/developerguide/custom-key-store-overview.html).
	// To find the ID of a custom key store, use the DescribeCustomKeyStores operation.
	CustomKeyStoreId *string

	// The length of the byte string.
	NumberOfBytes *int32

	noSmithyDocumentSerde
}

type GenerateRandomOutput struct {

	// The random byte string. When you use the HTTP API or the Amazon Web Services
	// CLI, the value is Base64-encoded. Otherwise, it is not Base64-encoded.
	Plaintext []byte

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGenerateRandomMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpGenerateRandom{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpGenerateRandom{}, middleware.After)
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
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGenerateRandom(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opGenerateRandom(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "kms",
		OperationName: "GenerateRandom",
	}
}
