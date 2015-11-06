#!/bin/bash

aws s3 ls s3://sourcegraph-release/src/ | ./src_version.py latest
