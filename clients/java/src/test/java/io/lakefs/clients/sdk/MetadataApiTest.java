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

import io.lakefs.clients.sdk.ApiException;
import io.lakefs.clients.sdk.model.Error;
import io.lakefs.clients.sdk.model.StorageURI;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * API tests for MetadataApi
 */
@Disabled
public class MetadataApiTest {

    private final MetadataApi api = new MetadataApi();

    /**
     * return URI to a meta-range file
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void getMetaRangeTest() throws ApiException {
        String repository = null;
        String metaRange = null;
        StorageURI response = api.getMetaRange(repository, metaRange)
                .execute();
        // TODO: test validations
    }

    /**
     * return URI to a range file
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void getRangeTest() throws ApiException {
        String repository = null;
        String range = null;
        StorageURI response = api.getRange(repository, range)
                .execute();
        // TODO: test validations
    }

}
