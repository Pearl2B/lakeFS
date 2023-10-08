# coding: utf-8

"""
    lakeFS API

    lakeFS HTTP API

    The version of the OpenAPI document: 0.1.0
    Contact: services@treeverse.io
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


from __future__ import annotations
import pprint
import re  # noqa: F401
import json


from typing import Optional
from pydantic import BaseModel, Field, StrictInt, StrictStr

class ObjectError(BaseModel):
    """
    ObjectError
    """
    status_code: StrictInt = Field(..., description="HTTP status code associated for operation on path")
    message: StrictStr = Field(..., description="short message explaining status_code")
    path: Optional[StrictStr] = Field(None, description="affected path")
    __properties = ["status_code", "message", "path"]

    class Config:
        """Pydantic configuration"""
        allow_population_by_field_name = True
        validate_assignment = True

    def to_str(self) -> str:
        """Returns the string representation of the model using alias"""
        return pprint.pformat(self.dict(by_alias=True))

    def to_json(self) -> str:
        """Returns the JSON representation of the model using alias"""
        return json.dumps(self.to_dict())

    @classmethod
    def from_json(cls, json_str: str) -> ObjectError:
        """Create an instance of ObjectError from a JSON string"""
        return cls.from_dict(json.loads(json_str))

    def to_dict(self):
        """Returns the dictionary representation of the model using alias"""
        _dict = self.dict(by_alias=True,
                          exclude={
                          },
                          exclude_none=True)
        return _dict

    @classmethod
    def from_dict(cls, obj: dict) -> ObjectError:
        """Create an instance of ObjectError from a dict"""
        if obj is None:
            return None

        if not isinstance(obj, dict):
            return ObjectError.parse_obj(obj)

        _obj = ObjectError.parse_obj({
            "status_code": obj.get("status_code"),
            "message": obj.get("message"),
            "path": obj.get("path")
        })
        return _obj

