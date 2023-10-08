/*
 * lakeFS API
 * lakeFS HTTP API
 *
 * The version of the OpenAPI document: 0.1.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */


package io.lakefs.clients.sdk;

import io.lakefs.clients.sdk.ApiCallback;
import io.lakefs.clients.sdk.ApiClient;
import io.lakefs.clients.sdk.ApiException;
import io.lakefs.clients.sdk.ApiResponse;
import io.lakefs.clients.sdk.Configuration;
import io.lakefs.clients.sdk.Pair;
import io.lakefs.clients.sdk.ProgressRequestBody;
import io.lakefs.clients.sdk.ProgressResponseBody;

import com.google.gson.reflect.TypeToken;

import java.io.IOException;


import io.lakefs.clients.sdk.model.Error;
import io.lakefs.clients.sdk.model.StorageURI;

import java.lang.reflect.Type;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class MetadataApi {
    private ApiClient localVarApiClient;
    private int localHostIndex;
    private String localCustomBaseUrl;

    public MetadataApi() {
        this(Configuration.getDefaultApiClient());
    }

    public MetadataApi(ApiClient apiClient) {
        this.localVarApiClient = apiClient;
    }

    public ApiClient getApiClient() {
        return localVarApiClient;
    }

    public void setApiClient(ApiClient apiClient) {
        this.localVarApiClient = apiClient;
    }

    public int getHostIndex() {
        return localHostIndex;
    }

    public void setHostIndex(int hostIndex) {
        this.localHostIndex = hostIndex;
    }

    public String getCustomBaseUrl() {
        return localCustomBaseUrl;
    }

    public void setCustomBaseUrl(String customBaseUrl) {
        this.localCustomBaseUrl = customBaseUrl;
    }

    private okhttp3.Call getMetaRangeCall(String repository, String metaRange, final ApiCallback _callback) throws ApiException {
        String basePath = null;
        // Operation Servers
        String[] localBasePaths = new String[] {  };

        // Determine Base Path to Use
        if (localCustomBaseUrl != null){
            basePath = localCustomBaseUrl;
        } else if ( localBasePaths.length > 0 ) {
            basePath = localBasePaths[localHostIndex];
        } else {
            basePath = null;
        }

        Object localVarPostBody = null;

        // create path and map variables
        String localVarPath = "/repositories/{repository}/metadata/meta_range/{meta_range}"
            .replace("{" + "repository" + "}", localVarApiClient.escapeString(repository.toString()))
            .replace("{" + "meta_range" + "}", localVarApiClient.escapeString(metaRange.toString()));

        List<Pair> localVarQueryParams = new ArrayList<Pair>();
        List<Pair> localVarCollectionQueryParams = new ArrayList<Pair>();
        Map<String, String> localVarHeaderParams = new HashMap<String, String>();
        Map<String, String> localVarCookieParams = new HashMap<String, String>();
        Map<String, Object> localVarFormParams = new HashMap<String, Object>();

        final String[] localVarAccepts = {
            "application/json"
        };
        final String localVarAccept = localVarApiClient.selectHeaderAccept(localVarAccepts);
        if (localVarAccept != null) {
            localVarHeaderParams.put("Accept", localVarAccept);
        }

        final String[] localVarContentTypes = {
        };
        final String localVarContentType = localVarApiClient.selectHeaderContentType(localVarContentTypes);
        if (localVarContentType != null) {
            localVarHeaderParams.put("Content-Type", localVarContentType);
        }

        String[] localVarAuthNames = new String[] { "basic_auth", "cookie_auth", "oidc_auth", "saml_auth", "jwt_token" };
        return localVarApiClient.buildCall(basePath, localVarPath, "GET", localVarQueryParams, localVarCollectionQueryParams, localVarPostBody, localVarHeaderParams, localVarCookieParams, localVarFormParams, localVarAuthNames, _callback);
    }

    @SuppressWarnings("rawtypes")
    private okhttp3.Call getMetaRangeValidateBeforeCall(String repository, String metaRange, final ApiCallback _callback) throws ApiException {
        // verify the required parameter 'repository' is set
        if (repository == null) {
            throw new ApiException("Missing the required parameter 'repository' when calling getMetaRange(Async)");
        }

        // verify the required parameter 'metaRange' is set
        if (metaRange == null) {
            throw new ApiException("Missing the required parameter 'metaRange' when calling getMetaRange(Async)");
        }

        return getMetaRangeCall(repository, metaRange, _callback);

    }


    private ApiResponse<StorageURI> getMetaRangeWithHttpInfo(String repository, String metaRange) throws ApiException {
        okhttp3.Call localVarCall = getMetaRangeValidateBeforeCall(repository, metaRange, null);
        Type localVarReturnType = new TypeToken<StorageURI>(){}.getType();
        return localVarApiClient.execute(localVarCall, localVarReturnType);
    }

    private okhttp3.Call getMetaRangeAsync(String repository, String metaRange, final ApiCallback<StorageURI> _callback) throws ApiException {

        okhttp3.Call localVarCall = getMetaRangeValidateBeforeCall(repository, metaRange, _callback);
        Type localVarReturnType = new TypeToken<StorageURI>(){}.getType();
        localVarApiClient.executeAsync(localVarCall, localVarReturnType, _callback);
        return localVarCall;
    }

    public class APIgetMetaRangeRequest {
        private final String repository;
        private final String metaRange;

        private APIgetMetaRangeRequest(String repository, String metaRange) {
            this.repository = repository;
            this.metaRange = metaRange;
        }

        /**
         * Build call for getMetaRange
         * @param _callback ApiCallback API callback
         * @return Call to execute
         * @throws ApiException If fail to serialize the request body object
         * @http.response.details
         <table summary="Response Details" border="1">
            <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
            <tr><td> 200 </td><td> meta-range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
            <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
            <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
            <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
            <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
         </table>
         */
        public okhttp3.Call buildCall(final ApiCallback _callback) throws ApiException {
            return getMetaRangeCall(repository, metaRange, _callback);
        }

        /**
         * Execute getMetaRange request
         * @return StorageURI
         * @throws ApiException If fail to call the API, e.g. server error or cannot deserialize the response body
         * @http.response.details
         <table summary="Response Details" border="1">
            <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
            <tr><td> 200 </td><td> meta-range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
            <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
            <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
            <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
            <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
         </table>
         */
        public StorageURI execute() throws ApiException {
            ApiResponse<StorageURI> localVarResp = getMetaRangeWithHttpInfo(repository, metaRange);
            return localVarResp.getData();
        }

        /**
         * Execute getMetaRange request with HTTP info returned
         * @return ApiResponse&lt;StorageURI&gt;
         * @throws ApiException If fail to call the API, e.g. server error or cannot deserialize the response body
         * @http.response.details
         <table summary="Response Details" border="1">
            <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
            <tr><td> 200 </td><td> meta-range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
            <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
            <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
            <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
            <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
         </table>
         */
        public ApiResponse<StorageURI> executeWithHttpInfo() throws ApiException {
            return getMetaRangeWithHttpInfo(repository, metaRange);
        }

        /**
         * Execute getMetaRange request (asynchronously)
         * @param _callback The callback to be executed when the API call finishes
         * @return The request call
         * @throws ApiException If fail to process the API call, e.g. serializing the request body object
         * @http.response.details
         <table summary="Response Details" border="1">
            <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
            <tr><td> 200 </td><td> meta-range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
            <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
            <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
            <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
            <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
         </table>
         */
        public okhttp3.Call executeAsync(final ApiCallback<StorageURI> _callback) throws ApiException {
            return getMetaRangeAsync(repository, metaRange, _callback);
        }
    }

    /**
     * return URI to a meta-range file
     * 
     * @param repository  (required)
     * @param metaRange  (required)
     * @return APIgetMetaRangeRequest
     * @http.response.details
     <table summary="Response Details" border="1">
        <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
        <tr><td> 200 </td><td> meta-range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
        <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
        <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
        <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
        <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
     </table>
     */
    public APIgetMetaRangeRequest getMetaRange(String repository, String metaRange) {
        return new APIgetMetaRangeRequest(repository, metaRange);
    }
    private okhttp3.Call getRangeCall(String repository, String range, final ApiCallback _callback) throws ApiException {
        String basePath = null;
        // Operation Servers
        String[] localBasePaths = new String[] {  };

        // Determine Base Path to Use
        if (localCustomBaseUrl != null){
            basePath = localCustomBaseUrl;
        } else if ( localBasePaths.length > 0 ) {
            basePath = localBasePaths[localHostIndex];
        } else {
            basePath = null;
        }

        Object localVarPostBody = null;

        // create path and map variables
        String localVarPath = "/repositories/{repository}/metadata/range/{range}"
            .replace("{" + "repository" + "}", localVarApiClient.escapeString(repository.toString()))
            .replace("{" + "range" + "}", localVarApiClient.escapeString(range.toString()));

        List<Pair> localVarQueryParams = new ArrayList<Pair>();
        List<Pair> localVarCollectionQueryParams = new ArrayList<Pair>();
        Map<String, String> localVarHeaderParams = new HashMap<String, String>();
        Map<String, String> localVarCookieParams = new HashMap<String, String>();
        Map<String, Object> localVarFormParams = new HashMap<String, Object>();

        final String[] localVarAccepts = {
            "application/json"
        };
        final String localVarAccept = localVarApiClient.selectHeaderAccept(localVarAccepts);
        if (localVarAccept != null) {
            localVarHeaderParams.put("Accept", localVarAccept);
        }

        final String[] localVarContentTypes = {
        };
        final String localVarContentType = localVarApiClient.selectHeaderContentType(localVarContentTypes);
        if (localVarContentType != null) {
            localVarHeaderParams.put("Content-Type", localVarContentType);
        }

        String[] localVarAuthNames = new String[] { "basic_auth", "cookie_auth", "oidc_auth", "saml_auth", "jwt_token" };
        return localVarApiClient.buildCall(basePath, localVarPath, "GET", localVarQueryParams, localVarCollectionQueryParams, localVarPostBody, localVarHeaderParams, localVarCookieParams, localVarFormParams, localVarAuthNames, _callback);
    }

    @SuppressWarnings("rawtypes")
    private okhttp3.Call getRangeValidateBeforeCall(String repository, String range, final ApiCallback _callback) throws ApiException {
        // verify the required parameter 'repository' is set
        if (repository == null) {
            throw new ApiException("Missing the required parameter 'repository' when calling getRange(Async)");
        }

        // verify the required parameter 'range' is set
        if (range == null) {
            throw new ApiException("Missing the required parameter 'range' when calling getRange(Async)");
        }

        return getRangeCall(repository, range, _callback);

    }


    private ApiResponse<StorageURI> getRangeWithHttpInfo(String repository, String range) throws ApiException {
        okhttp3.Call localVarCall = getRangeValidateBeforeCall(repository, range, null);
        Type localVarReturnType = new TypeToken<StorageURI>(){}.getType();
        return localVarApiClient.execute(localVarCall, localVarReturnType);
    }

    private okhttp3.Call getRangeAsync(String repository, String range, final ApiCallback<StorageURI> _callback) throws ApiException {

        okhttp3.Call localVarCall = getRangeValidateBeforeCall(repository, range, _callback);
        Type localVarReturnType = new TypeToken<StorageURI>(){}.getType();
        localVarApiClient.executeAsync(localVarCall, localVarReturnType, _callback);
        return localVarCall;
    }

    public class APIgetRangeRequest {
        private final String repository;
        private final String range;

        private APIgetRangeRequest(String repository, String range) {
            this.repository = repository;
            this.range = range;
        }

        /**
         * Build call for getRange
         * @param _callback ApiCallback API callback
         * @return Call to execute
         * @throws ApiException If fail to serialize the request body object
         * @http.response.details
         <table summary="Response Details" border="1">
            <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
            <tr><td> 200 </td><td> range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
            <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
            <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
            <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
            <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
         </table>
         */
        public okhttp3.Call buildCall(final ApiCallback _callback) throws ApiException {
            return getRangeCall(repository, range, _callback);
        }

        /**
         * Execute getRange request
         * @return StorageURI
         * @throws ApiException If fail to call the API, e.g. server error or cannot deserialize the response body
         * @http.response.details
         <table summary="Response Details" border="1">
            <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
            <tr><td> 200 </td><td> range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
            <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
            <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
            <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
            <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
         </table>
         */
        public StorageURI execute() throws ApiException {
            ApiResponse<StorageURI> localVarResp = getRangeWithHttpInfo(repository, range);
            return localVarResp.getData();
        }

        /**
         * Execute getRange request with HTTP info returned
         * @return ApiResponse&lt;StorageURI&gt;
         * @throws ApiException If fail to call the API, e.g. server error or cannot deserialize the response body
         * @http.response.details
         <table summary="Response Details" border="1">
            <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
            <tr><td> 200 </td><td> range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
            <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
            <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
            <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
            <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
         </table>
         */
        public ApiResponse<StorageURI> executeWithHttpInfo() throws ApiException {
            return getRangeWithHttpInfo(repository, range);
        }

        /**
         * Execute getRange request (asynchronously)
         * @param _callback The callback to be executed when the API call finishes
         * @return The request call
         * @throws ApiException If fail to process the API call, e.g. serializing the request body object
         * @http.response.details
         <table summary="Response Details" border="1">
            <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
            <tr><td> 200 </td><td> range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
            <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
            <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
            <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
            <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
         </table>
         */
        public okhttp3.Call executeAsync(final ApiCallback<StorageURI> _callback) throws ApiException {
            return getRangeAsync(repository, range, _callback);
        }
    }

    /**
     * return URI to a range file
     * 
     * @param repository  (required)
     * @param range  (required)
     * @return APIgetRangeRequest
     * @http.response.details
     <table summary="Response Details" border="1">
        <tr><td> Status Code </td><td> Description </td><td> Response Headers </td></tr>
        <tr><td> 200 </td><td> range URI </td><td>  * Location - redirect to S3 <br>  </td></tr>
        <tr><td> 401 </td><td> Unauthorized </td><td>  -  </td></tr>
        <tr><td> 404 </td><td> Resource Not Found </td><td>  -  </td></tr>
        <tr><td> 420 </td><td> too many requests </td><td>  -  </td></tr>
        <tr><td> 0 </td><td> Internal Server Error </td><td>  -  </td></tr>
     </table>
     */
    public APIgetRangeRequest getRange(String repository, String range) {
        return new APIgetRangeRequest(repository, range);
    }
}