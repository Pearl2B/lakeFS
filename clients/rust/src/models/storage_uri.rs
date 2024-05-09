/*
 * lakeFS API
 *
 * lakeFS HTTP API
 *
 * The version of the OpenAPI document: 1.0.0
 * Contact: services@treeverse.io
 * Generated by: https://openapi-generator.tech
 */

use crate::models;

/// StorageUri : URI to a path in a storage provider (e.g. \"s3://bucket1/path/to/object\")
#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct StorageUri {
    #[serde(rename = "location")]
    pub location: String,
}

impl StorageUri {
    /// URI to a path in a storage provider (e.g. \"s3://bucket1/path/to/object\")
    pub fn new(location: String) -> StorageUri {
        StorageUri {
            location,
        }
    }
}
