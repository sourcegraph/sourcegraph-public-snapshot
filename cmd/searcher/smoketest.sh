#!/bin/bash

# A simple script to directly query the search service.

set -ex
curl 'http://localhost:3181?Repo=github.com/gorilla/mux&Commit=599cba5e7b6137d46ddf58fb1765f5d928e69604&Pattern=Router'
