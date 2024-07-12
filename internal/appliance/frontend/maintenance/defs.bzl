"Constants for conventions used for React source files"

load("@aspect_rules_swc//swc:defs.bzl", "swc")
load("@bazel_skylib//lib:partial.bzl", "partial")

ASSET_PATTERNS = [
    "*.svg",
    "*.css",
]

SRC_PATTERNS = [
    "*.tsx",
    "*.ts",
]

# Filename conventions described at
# https://create-react-app.dev/docs/running-tests#filename-conventions
TEST_PATTERNS = [
    "*.test.tsx",
    "*.test.ts",
    "*.spec.tsx",
    "*.spec.ts",
]

# Partially-apply our (generated) .swcrc config file to the swc starlark function
TRANSPILER = partial.make(
    swc,
    swcrc = "//react:.swcrc",
)
