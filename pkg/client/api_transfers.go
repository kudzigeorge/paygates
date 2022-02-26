/*
 * Paygate API
 *
 * PayGate is a RESTful API enabling first-party Automated Clearing House ([ACH](https://en.wikipedia.org/wiki/Automated_Clearing_House)) transfers to be created without a deep understanding of a full NACHA file specification. First-party transfers initiate at an Originating Depository Financial Institution (ODFI) and are sent off to other Financial Institutions.  An organization is a value used to isolate models from each other. This can be set to a \"user ID\" from your authentication service or any value your system has to identify.  There are also [admin endpoints](https://moov-io.github.io/paygate/admin/) for back-office operations.
 *
 * API version: v1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package client

import (
	_context "context"
	"github.com/antihax/optional"
	_ioutil "io/ioutil"
	_nethttp "net/http"
	_neturl "net/url"
	"strings"
)

// Linger please
var (
	_ _context.Context
)

// TransfersApiService TransfersApi service
type TransfersApiService service

// AddTransferOpts Optional parameters for the method 'AddTransfer'
type AddTransferOpts struct {
	XIdempotencyKey optional.String
	XRequestID      optional.String
}

/*
AddTransfer Create Transfer
Create a new transfer between a Source and a Destination. Transfers can only be modified in the pending status.
 * @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param xOrganization Value used to separate and identify models
 * @param createTransfer
 * @param optional nil or *AddTransferOpts - Optional Parameters:
 * @param "XIdempotencyKey" (optional.String) -  Idempotent key in the header which expires after 24 hours. These strings should contain enough entropy for to not collide with each other in your requests.
 * @param "XRequestID" (optional.String) -  Optional requestID allows application developer to trace requests through the systems logs
@return Transfer
*/
func (a *TransfersApiService) AddTransfer(ctx _context.Context, xOrganization string, createTransfer CreateTransfer, localVarOptionals *AddTransferOpts) (Transfer, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodPost
		localVarPostBody     interface{}
		localVarFormFileName string
		localVarFileName     string
		localVarFileBytes    []byte
		localVarReturnValue  Transfer
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/transfers"
	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XIdempotencyKey.IsSet() {
		localVarHeaderParams["X-Idempotency-Key"] = parameterToString(localVarOptionals.XIdempotencyKey.Value(), "")
	}
	if localVarOptionals != nil && localVarOptionals.XRequestID.IsSet() {
		localVarHeaderParams["X-Request-ID"] = parameterToString(localVarOptionals.XRequestID.Value(), "")
	}
	localVarHeaderParams["X-Organization"] = parameterToString(xOrganization, "")
	// body params
	localVarPostBody = &createTransfer
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFormFileName, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v Error
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHTTPResponse, newErr
		}
		if localVarHTTPResponse.StatusCode == 412 {
			var v Error
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// DeleteTransferByIDOpts Optional parameters for the method 'DeleteTransferByID'
type DeleteTransferByIDOpts struct {
	XRequestID optional.String
}

/*
DeleteTransferByID Delete Transfer
Remove a transfer for the specified organization. Its status will be updated as transfer is processed. It is only possible to delete (recall) a Transfer before it has been released from the financial institution.
 * @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param transferID transferID to delete
 * @param xOrganization Value used to separate and identify models
 * @param optional nil or *DeleteTransferByIDOpts - Optional Parameters:
 * @param "XRequestID" (optional.String) -  Optional requestID allows application developer to trace requests through the systems logs
*/
func (a *TransfersApiService) DeleteTransferByID(ctx _context.Context, transferID string, xOrganization string, localVarOptionals *DeleteTransferByIDOpts) (*_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodDelete
		localVarPostBody     interface{}
		localVarFormFileName string
		localVarFileName     string
		localVarFileBytes    []byte
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/transfers/{transferID}"
	localVarPath = strings.Replace(localVarPath, "{"+"transferID"+"}", _neturl.QueryEscape(parameterToString(transferID, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XRequestID.IsSet() {
		localVarHeaderParams["X-Request-ID"] = parameterToString(localVarOptionals.XRequestID.Value(), "")
	}
	localVarHeaderParams["X-Organization"] = parameterToString(xOrganization, "")
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFormFileName, localVarFileName, localVarFileBytes)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v Error
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

// GetTransferByIDOpts Optional parameters for the method 'GetTransferByID'
type GetTransferByIDOpts struct {
	XRequestID optional.String
}

/*
GetTransferByID Get Transfer
Get a Transfer object for the supplied organization
 * @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param transferID transferID to retrieve
 * @param xOrganization Value used to separate and identify models
 * @param optional nil or *GetTransferByIDOpts - Optional Parameters:
 * @param "XRequestID" (optional.String) -  Optional requestID allows application developer to trace requests through the systems logs
@return Transfer
*/
func (a *TransfersApiService) GetTransferByID(ctx _context.Context, transferID string, xOrganization string, localVarOptionals *GetTransferByIDOpts) (Transfer, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodGet
		localVarPostBody     interface{}
		localVarFormFileName string
		localVarFileName     string
		localVarFileBytes    []byte
		localVarReturnValue  Transfer
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/transfers/{transferID}"
	localVarPath = strings.Replace(localVarPath, "{"+"transferID"+"}", _neturl.QueryEscape(parameterToString(transferID, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XRequestID.IsSet() {
		localVarHeaderParams["X-Request-ID"] = parameterToString(localVarOptionals.XRequestID.Value(), "")
	}
	localVarHeaderParams["X-Organization"] = parameterToString(xOrganization, "")
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFormFileName, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

// GetTransfersOpts Optional parameters for the method 'GetTransfers'
type GetTransfersOpts struct {
	Skip            optional.Int32
	Count           optional.Int32
	Status          optional.Interface
	StartDate       optional.Time
	EndDate         optional.Time
	OrganizationIDs optional.String
	CustomerIDs     optional.String
	XRequestID      optional.String
}

/*
GetTransfers List Transfers
List all Transfers created for the given organization.
 * @param ctx _context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param xOrganization Value used to separate and identify models
 * @param optional nil or *GetTransfersOpts - Optional Parameters:
 * @param "Skip" (optional.Int32) -  The number of items to skip before starting to collect the result set
 * @param "Count" (optional.Int32) -  The number of items to return
 * @param "Status" (optional.Interface of TransferStatus) -  Return only Transfers in this TransferStatus
 * @param "StartDate" (optional.Time) -  Return Transfers that are scheduled for this date or later in ISO-8601 format YYYY-MM-DD. Can optionally be used with endDate to specify a date range.
 * @param "EndDate" (optional.Time) -  Return Transfers that are scheduled for this date or earlier in ISO-8601 format YYYY-MM-DD. Can optionally be used with startDate to specify a date range.
 * @param "OrganizationIDs" (optional.String) -  Comma separated list of organizationID values to return Transfer objects for.
 * @param "CustomerIDs" (optional.String) -  Comma separated list of customerID values to return Transfer objects for. A maximum of 25 IDs is allowed.
 * @param "XRequestID" (optional.String) -  Optional requestID allows application developer to trace requests through the systems logs
@return []Transfer
*/
func (a *TransfersApiService) GetTransfers(ctx _context.Context, xOrganization string, localVarOptionals *GetTransfersOpts) ([]Transfer, *_nethttp.Response, error) {
	var (
		localVarHTTPMethod   = _nethttp.MethodGet
		localVarPostBody     interface{}
		localVarFormFileName string
		localVarFileName     string
		localVarFileBytes    []byte
		localVarReturnValue  []Transfer
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/transfers"
	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := _neturl.Values{}
	localVarFormParams := _neturl.Values{}

	if localVarOptionals != nil && localVarOptionals.Skip.IsSet() {
		localVarQueryParams.Add("skip", parameterToString(localVarOptionals.Skip.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Count.IsSet() {
		localVarQueryParams.Add("count", parameterToString(localVarOptionals.Count.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Status.IsSet() {
		localVarQueryParams.Add("status", parameterToString(localVarOptionals.Status.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.StartDate.IsSet() {
		localVarQueryParams.Add("startDate", parameterToString(localVarOptionals.StartDate.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.EndDate.IsSet() {
		localVarQueryParams.Add("endDate", parameterToString(localVarOptionals.EndDate.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.OrganizationIDs.IsSet() {
		localVarQueryParams.Add("organizationIDs", parameterToString(localVarOptionals.OrganizationIDs.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.CustomerIDs.IsSet() {
		localVarQueryParams.Add("customerIDs", parameterToString(localVarOptionals.CustomerIDs.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if localVarOptionals != nil && localVarOptionals.XRequestID.IsSet() {
		localVarHeaderParams["X-Request-ID"] = parameterToString(localVarOptionals.XRequestID.Value(), "")
	}
	localVarHeaderParams["X-Organization"] = parameterToString(xOrganization, "")
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFormFileName, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(r)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := _ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		if localVarHTTPResponse.StatusCode == 400 {
			var v Error
			err = a.client.decode(&v, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHTTPResponse, newErr
			}
			newErr.model = v
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
