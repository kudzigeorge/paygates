# \TransfersApi

All URIs are relative to *http://localhost:9092*

Method | HTTP request | Description
------------- | ------------- | -------------
[**TriggerCutoffProcessing**](TransfersApi.md#TriggerCutoffProcessing) | **Put** /trigger-cutoff | Initiate cutoff processing
[**UpdateTransferStatus**](TransfersApi.md#UpdateTransferStatus) | **Put** /transfers/{transferId}/status | Update Transfer status



## TriggerCutoffProcessing

> TriggerCutoffProcessing(ctx, )

Initiate cutoff processing

Starts processing like it's a cutoff window approaching. This involves merging transfers into files, upload attempts, along with inbound file download processing.

### Required Parameters

This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTransferStatus

> UpdateTransferStatus(ctx, transferId, xOrganization, updateTransferStatus, optional)

Update Transfer status

Updates a Transfer status for a transferID.  PENDING transfers may be updated to: CANCELED or REVIEWABLE. REVIEWABLE transfers may be updated to: CANCELED or PENDING. 

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**transferId** | **string**| transferID that identifies the Transfer | 
**xOrganization** | **string**| Value used to separate and identify models | 
**updateTransferStatus** | [**UpdateTransferStatus**](UpdateTransferStatus.md)|  | 
 **optional** | ***UpdateTransferStatusOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UpdateTransferStatusOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **xRequestID** | **optional.String**| Optional requestID allows application developer to trace requests through the systems logs | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

