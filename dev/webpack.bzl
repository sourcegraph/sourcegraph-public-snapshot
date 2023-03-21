load("@aspect_rules_webpack//webpack:defs.bzl", _webpack_bundle = "webpack_bundle", _webpack_devserver = "webpack_devserver")

def webpack_bundle(name, **kwargs):
    _webpack_bundle(
        name = name,
        webpack = "//dev:webpack",
        **kwargs
    )

def webpack_devserver(name, **kwargs):
    _webpack_devserver(
        name = name,
        webpack = "//dev:webpack",
        **kwargs
    )
