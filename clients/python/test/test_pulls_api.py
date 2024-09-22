# coding: utf-8

"""
    lakeFS API

    lakeFS HTTP API

    The version of the OpenAPI document: 1.0.0
    Contact: services@treeverse.io
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


import unittest

import lakefs_sdk
from lakefs_sdk.api.pulls_api import PullsApi  # noqa: E501
from lakefs_sdk.rest import ApiException


class TestPullsApi(unittest.TestCase):
    """PullsApi unit test stubs"""

    def setUp(self):
        self.api = lakefs_sdk.api.pulls_api.PullsApi()  # noqa: E501

    def tearDown(self):
        pass

    def test_create_pull_request(self):
        """Test case for create_pull_request

        create pull request  # noqa: E501
        """
        pass

    def test_get_pull_request(self):
        """Test case for get_pull_request

        get pull request  # noqa: E501
        """
        pass

    def test_list_pull_requests(self):
        """Test case for list_pull_requests

        list pull requests  # noqa: E501
        """
        pass

    def test_update_pull_request(self):
        """Test case for update_pull_request

        update pull request  # noqa: E501
        """
        pass


if __name__ == '__main__':
    unittest.main()
