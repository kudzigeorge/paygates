# \ConfigurationApi

All URIs are relative to *http://localhost:8082*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetTransferConfiguration**](ConfigurationApi.md#GetTransferConfiguration) | **Get** /configuration/transfers | Get Configuration
[**UpdateTransferConfiguration**](ConfigurationApi.md#UpdateTransferConfiguration) | **Put** /configuration/transfers | Update Configuration



## GetTransferConfiguration

> OrganizationConfiguration GetTransferConfiguration(ctx, optional)

Get Configuration

Retrieve current config for the provided organization.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***GetTransferConfigurationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a GetTransferConfigurationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**OrganizationConfiguration**](OrganizationConfiguration.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTransferConfiguration

> OrganizationConfiguration UpdateTransferConfiguration(ctx, organizationConfiguration, optional)

Update Configuration

Update the config for the provided organization.

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**organizationConfiguration** | [**OrganizationConfiguration**](OrganizationConfiguration.md)|  | 
 **optional** | ***UpdateTransferConfigurationOpts** | optional parameters | nil if no parameters

### Optional Parameters

Optional parameters are passed through a pointer to a UpdateTransferConfigurationOpts struct


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **xOrganization** | **optional.String**| Value used to separate and identify models | 

### Return type

[**OrganizationConfiguration**](OrganizationConfiguration.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

