"""
    lakeFS API

    lakeFS HTTP API  # noqa: E501

    The version of the OpenAPI document: 1.0.0
    Contact: services@treeverse.io
    Generated by: https://openapi-generator.tech
"""


import sys
import unittest

import lakefs_client
from lakefs_client.model.object_user_metadata import ObjectUserMetadata
globals()['ObjectUserMetadata'] = ObjectUserMetadata
from lakefs_client.model.update_object_user_metadata import UpdateObjectUserMetadata


class TestUpdateObjectUserMetadata(unittest.TestCase):
    """UpdateObjectUserMetadata unit test stubs"""

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def testUpdateObjectUserMetadata(self):
        """Test UpdateObjectUserMetadata"""
        # FIXME: construct object with mandatory attributes with example values
        # model = UpdateObjectUserMetadata()  # noqa: E501
        pass


if __name__ == '__main__':
    unittest.main()