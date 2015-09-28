# This file is used for testing only

# To perform the test, run `cmake .` at the root of the project tree followed
# by `ctest .`

cmake_minimum_required(VERSION 2.6)

# Do not check any compiler
project(editorconfig-core-js NONE)

set(NODE node CACHE STRING "Node.js command")

enable_testing()

set(EDITORCONFIG_CMD ${NODE} ${PROJECT_SOURCE_DIR}/bin/editorconfig)
add_subdirectory(tests)
