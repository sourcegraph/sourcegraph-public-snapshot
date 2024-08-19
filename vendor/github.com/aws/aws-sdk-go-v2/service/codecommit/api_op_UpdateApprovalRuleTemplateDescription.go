// Code generated by smithy-go-codegen DO NOT EDIT.

package codecommit

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/codecommit/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Updates the description for a specified approval rule template.
func (c *Client) UpdateApprovalRuleTemplateDescription(ctx context.Context, params *UpdateApprovalRuleTemplateDescriptionInput, optFns ...func(*Options)) (*UpdateApprovalRuleTemplateDescriptionOutput, error) {
	if params == nil {
		params = &UpdateApprovalRuleTemplateDescriptionInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "UpdateApprovalRuleTemplateDescription", params, optFns, c.addOperationUpdateApprovalRuleTemplateDescriptionMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*UpdateApprovalRuleTemplateDescriptionOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type UpdateApprovalRuleTemplateDescriptionInput struct {

	// The updated description of the approval rule template.
	//
	// This member is required.
	ApprovalRuleTemplateDescription *string

	// The name of the template for which you want to update the description.
	//
	// This member is required.
	ApprovalRuleTemplateName *string

	noSmithyDocumentSerde
}

type UpdateApprovalRuleTemplateDescriptionOutput struct {

	// The structure and content of the updated approval rule template.
	//
	// This member is required.
	ApprovalRuleTemplate *types.ApprovalRuleTemplate

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationUpdateApprovalRuleTemplateDescriptionMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpUpdateApprovalRuleTemplateDescription{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpUpdateApprovalRuleTemplateDescription{}, middleware.After)
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
	if err = addOpUpdateApprovalRuleTemplateDescriptionValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opUpdateApprovalRuleTemplateDescription(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opUpdateApprovalRuleTemplateDescription(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "codecommit",
		OperationName: "UpdateApprovalRuleTemplateDescription",
	}
}
