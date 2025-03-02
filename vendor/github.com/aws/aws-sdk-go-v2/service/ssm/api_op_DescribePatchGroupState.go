// Code generated by smithy-go-codegen DO NOT EDIT.

package ssm

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Returns high-level aggregated patch compliance state information for a patch
// group.
func (c *Client) DescribePatchGroupState(ctx context.Context, params *DescribePatchGroupStateInput, optFns ...func(*Options)) (*DescribePatchGroupStateOutput, error) {
	if params == nil {
		params = &DescribePatchGroupStateInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DescribePatchGroupState", params, optFns, c.addOperationDescribePatchGroupStateMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DescribePatchGroupStateOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DescribePatchGroupStateInput struct {

	// The name of the patch group whose patch snapshot should be retrieved.
	//
	// This member is required.
	PatchGroup *string

	noSmithyDocumentSerde
}

type DescribePatchGroupStateOutput struct {

	// The number of managed nodes in the patch group.
	Instances int32

	// The number of managed nodes where patches that are specified as Critical for
	// compliance reporting in the patch baseline aren't installed. These patches might
	// be missing, have failed installation, were rejected, or were installed but
	// awaiting a required managed node reboot. The status of these managed nodes is
	// NON_COMPLIANT .
	InstancesWithCriticalNonCompliantPatches *int32

	// The number of managed nodes with patches from the patch baseline that failed to
	// install.
	InstancesWithFailedPatches int32

	// The number of managed nodes with patches installed that aren't defined in the
	// patch baseline.
	InstancesWithInstalledOtherPatches int32

	// The number of managed nodes with installed patches.
	InstancesWithInstalledPatches int32

	// The number of managed nodes with patches installed by Patch Manager that
	// haven't been rebooted after the patch installation. The status of these managed
	// nodes is NON_COMPLIANT .
	InstancesWithInstalledPendingRebootPatches *int32

	// The number of managed nodes with patches installed that are specified in a
	// RejectedPatches list. Patches with a status of INSTALLED_REJECTED were
	// typically installed before they were added to a RejectedPatches list.
	//
	// If ALLOW_AS_DEPENDENCY is the specified option for RejectedPatchesAction , the
	// value of InstancesWithInstalledRejectedPatches will always be 0 (zero).
	InstancesWithInstalledRejectedPatches *int32

	// The number of managed nodes with missing patches from the patch baseline.
	InstancesWithMissingPatches int32

	// The number of managed nodes with patches that aren't applicable.
	InstancesWithNotApplicablePatches int32

	// The number of managed nodes with patches installed that are specified as other
	// than Critical or Security but aren't compliant with the patch baseline. The
	// status of these managed nodes is NON_COMPLIANT .
	InstancesWithOtherNonCompliantPatches *int32

	// The number of managed nodes where patches that are specified as Security in a
	// patch advisory aren't installed. These patches might be missing, have failed
	// installation, were rejected, or were installed but awaiting a required managed
	// node reboot. The status of these managed nodes is NON_COMPLIANT .
	InstancesWithSecurityNonCompliantPatches *int32

	// The number of managed nodes with NotApplicable patches beyond the supported
	// limit, which aren't reported by name to Inventory. Inventory is a tool in Amazon
	// Web Services Systems Manager.
	InstancesWithUnreportedNotApplicablePatches *int32

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationDescribePatchGroupStateMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpDescribePatchGroupState{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpDescribePatchGroupState{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "DescribePatchGroupState"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = addClientRequestID(stack); err != nil {
		return err
	}
	if err = addComputeContentLength(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = addComputePayloadSHA256(stack); err != nil {
		return err
	}
	if err = addRetry(stack, options); err != nil {
		return err
	}
	if err = addRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = addRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addSpanRetryLoop(stack, options); err != nil {
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
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addTimeOffsetBuild(stack, c); err != nil {
		return err
	}
	if err = addUserAgentRetryMode(stack, options); err != nil {
		return err
	}
	if err = addCredentialSource(stack, options); err != nil {
		return err
	}
	if err = addOpDescribePatchGroupStateValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDescribePatchGroupState(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRecursionDetection(stack); err != nil {
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
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	if err = addSpanInitializeStart(stack); err != nil {
		return err
	}
	if err = addSpanInitializeEnd(stack); err != nil {
		return err
	}
	if err = addSpanBuildRequestStart(stack); err != nil {
		return err
	}
	if err = addSpanBuildRequestEnd(stack); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opDescribePatchGroupState(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "DescribePatchGroupState",
	}
}
