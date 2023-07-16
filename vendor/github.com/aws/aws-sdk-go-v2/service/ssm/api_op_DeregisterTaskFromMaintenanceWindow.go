// Code generated by smithy-go-codegen DO NOT EDIT.

package ssm

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Removes a task from a maintenance window.
func (c *Client) DeregisterTaskFromMaintenanceWindow(ctx context.Context, params *DeregisterTaskFromMaintenanceWindowInput, optFns ...func(*Options)) (*DeregisterTaskFromMaintenanceWindowOutput, error) {
	if params == nil {
		params = &DeregisterTaskFromMaintenanceWindowInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DeregisterTaskFromMaintenanceWindow", params, optFns, c.addOperationDeregisterTaskFromMaintenanceWindowMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DeregisterTaskFromMaintenanceWindowOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DeregisterTaskFromMaintenanceWindowInput struct {

	// The ID of the maintenance window the task should be removed from.
	//
	// This member is required.
	WindowId *string

	// The ID of the task to remove from the maintenance window.
	//
	// This member is required.
	WindowTaskId *string

	noSmithyDocumentSerde
}

type DeregisterTaskFromMaintenanceWindowOutput struct {

	// The ID of the maintenance window the task was removed from.
	WindowId *string

	// The ID of the task removed from the maintenance window.
	WindowTaskId *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationDeregisterTaskFromMaintenanceWindowMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpDeregisterTaskFromMaintenanceWindow{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpDeregisterTaskFromMaintenanceWindow{}, middleware.After)
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
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpDeregisterTaskFromMaintenanceWindowValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDeregisterTaskFromMaintenanceWindow(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecursionDetection(stack); err != nil {
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

func newServiceMetadataMiddleware_opDeregisterTaskFromMaintenanceWindow(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "ssm",
		OperationName: "DeregisterTaskFromMaintenanceWindow",
	}
}
