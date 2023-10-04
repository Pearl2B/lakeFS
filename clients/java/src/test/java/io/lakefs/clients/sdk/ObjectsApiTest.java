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
import java.io.File;
import io.lakefs.clients.sdk.model.ObjectCopyCreation;
import io.lakefs.clients.sdk.model.ObjectErrorList;
import io.lakefs.clients.sdk.model.ObjectStats;
import io.lakefs.clients.sdk.model.ObjectStatsList;
import io.lakefs.clients.sdk.model.PathList;
import io.lakefs.clients.sdk.model.UnderlyingObjectProperties;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * API tests for ObjectsApi
 */
@Disabled
public class ObjectsApiTest {

    private final ObjectsApi api = new ObjectsApi();

    /**
     * create a copy of an object
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void copyObjectTest() throws ApiException {
        String repository = null;
        String branch = null;
        String destPath = null;
        ObjectCopyCreation objectCopyCreation = null;
        ObjectStats response = api.copyObject(repository, branch, destPath, objectCopyCreation)
                .execute();
        // TODO: test validations
    }

    /**
     * delete object. Missing objects will not return a NotFound error.
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void deleteObjectTest() throws ApiException {
        String repository = null;
        String branch = null;
        String path = null;
        api.deleteObject(repository, branch, path)
                .execute();
        // TODO: test validations
    }

    /**
     * delete objects. Missing objects will not return a NotFound error.
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void deleteObjectsTest() throws ApiException {
        String repository = null;
        String branch = null;
        PathList pathList = null;
        ObjectErrorList response = api.deleteObjects(repository, branch, pathList)
                .execute();
        // TODO: test validations
    }

    /**
     * get object content
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void getObjectTest() throws ApiException {
        String repository = null;
        String ref = null;
        String path = null;
        String range = null;
        Boolean presign = null;
        File response = api.getObject(repository, ref, path)
                .range(range)
                .presign(presign)
                .execute();
        // TODO: test validations
    }

    /**
     * get object properties on underlying storage
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void getUnderlyingPropertiesTest() throws ApiException {
        String repository = null;
        String ref = null;
        String path = null;
        UnderlyingObjectProperties response = api.getUnderlyingProperties(repository, ref, path)
                .execute();
        // TODO: test validations
    }

    /**
     * check if object exists
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void headObjectTest() throws ApiException {
        String repository = null;
        String ref = null;
        String path = null;
        String range = null;
        api.headObject(repository, ref, path)
                .range(range)
                .execute();
        // TODO: test validations
    }

    /**
     * list objects under a given prefix
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void listObjectsTest() throws ApiException {
        String repository = null;
        String ref = null;
        Boolean userMetadata = null;
        Boolean presign = null;
        String after = null;
        Integer amount = null;
        String delimiter = null;
        String prefix = null;
        ObjectStatsList response = api.listObjects(repository, ref)
                .userMetadata(userMetadata)
                .presign(presign)
                .after(after)
                .amount(amount)
                .delimiter(delimiter)
                .prefix(prefix)
                .execute();
        // TODO: test validations
    }

    /**
     * get object metadata
     *
     * @throws ApiException if the Api call fails
     */
    @Test
    public void statObjectTest() throws ApiException {
        String repository = null;
        String ref = null;
        String path = null;
        Boolean userMetadata = null;
        Boolean presign = null;
        ObjectStats response = api.statObject(repository, ref, path)
                .userMetadata(userMetadata)
                .presign(presign)
                .execute();
        // TODO: test validations
    }

    /**
     * @throws ApiException if the Api call fails
     */
    @Test
    public void uploadObjectTest() throws ApiException {
        String repository = null;
        String branch = null;
        String path = null;
        String storageClass = null;
        String ifNoneMatch = null;
        File content = null;
        ObjectStats response = api.uploadObject(repository, branch, path)
                .storageClass(storageClass)
                .ifNoneMatch(ifNoneMatch)
                .content(content)
                .execute();
        // TODO: test validations
    }

}
