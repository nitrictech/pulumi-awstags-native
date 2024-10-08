# coding=utf-8
# *** WARNING: this file was generated by pulumi-language-python. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

from . import _utilities
import typing
# Export this package's modules as members:
from .provider import *

# Make subpackages available:
if typing.TYPE_CHECKING:
    import pulumi_awstags.aws as __aws
    aws = __aws
else:
    aws = _utilities.lazy_import('pulumi_awstags.aws')

_utilities.register(
    resource_modules="""
[
 {
  "pkg": "awstags",
  "mod": "aws",
  "fqn": "pulumi_awstags.aws",
  "classes": {
   "awstags:aws:ResourceTag": "ResourceTag"
  }
 }
]
""",
    resource_packages="""
[
 {
  "pkg": "awstags",
  "token": "pulumi:providers:awstags",
  "fqn": "pulumi_awstags",
  "class": "Provider"
 }
]
"""
)
