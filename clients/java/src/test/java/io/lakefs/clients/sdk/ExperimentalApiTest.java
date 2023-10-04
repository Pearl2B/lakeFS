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
import io.lakefs.clients.sdk.model.OTFDiffs;
import io.lakefs.clients.sdk.model.OtfDiffList;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * API tests for ExperimentalApi
 */
@Disabled
public class ExperimentalApiTest {

    private final ExperimentalApi api = new ExperimentalApi();

    /**
     * get the available Open Table Format diffs
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void getOtfDiffsTest() throws ApiException {
        OTFDiffs response = api.getOtfDiffs()
                .execute();
        // TODO: test validations
    }

    /**
     * perform otf diff
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void otfDiffTest() throws ApiException {
        String repository = null;
        String leftRef = null;
        String rightRef = null;
        String tablePath = null;
        String type = null;
        OtfDiffList response = api.otfDiff(repository, leftRef, rightRef, tablePath, type)
                .execute();
        // TODO: test validations
    }

}
