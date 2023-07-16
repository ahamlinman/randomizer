// Code generated by smithy-go-codegen DO NOT EDIT.

package ssm

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Reconnects a session to a managed node after it has been disconnected.
// Connections can be resumed for disconnected sessions, but not terminated
// sessions. This command is primarily for use by client machines to automatically
// reconnect during intermittent network issues. It isn't intended for any other
// use.
func (c *Client) ResumeSession(ctx context.Context, params *ResumeSessionInput, optFns ...func(*Options)) (*ResumeSessionOutput, error) {
	if params == nil {
		params = &ResumeSessionInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ResumeSession", params, optFns, c.addOperationResumeSessionMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ResumeSessionOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ResumeSessionInput struct {

	// The ID of the disconnected session to resume.
	//
	// This member is required.
	SessionId *string

	noSmithyDocumentSerde
}

type ResumeSessionOutput struct {

	// The ID of the session.
	SessionId *string

	// A URL back to SSM Agent on the managed node that the Session Manager client
	// uses to send commands and receive output from the managed node. Format:
	// wss://ssmmessages.region.amazonaws.com/v1/data-channel/session-id?stream=(input|output)
	// . region represents the Region identifier for an Amazon Web Services Region
	// supported by Amazon Web Services Systems Manager, such as us-east-2 for the US
	// East (Ohio) Region. For a list of supported region values, see the Region column
	// in Systems Manager service endpoints (https://docs.aws.amazon.com/general/latest/gr/ssm.html#ssm_region)
	// in the Amazon Web Services General Reference. session-id represents the ID of a
	// Session Manager session, such as 1a2b3c4dEXAMPLE .
	StreamUrl *string

	// An encrypted token value containing session and caller information. Used to
	// authenticate the connection to the managed node.
	TokenValue *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationResumeSessionMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpResumeSession{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpResumeSession{}, middleware.After)
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
	if err = addOpResumeSessionValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opResumeSession(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opResumeSession(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "ssm",
		OperationName: "ResumeSession",
	}
}
