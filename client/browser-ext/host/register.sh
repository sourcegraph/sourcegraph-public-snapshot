#!/bin/bash

cd $(dirname $0)

mkdir -p /Library/Google/Chrome/NativeMessagingHosts
cp com.sourcegraph.browser_ext_host.json /Library/Google/Chrome/NativeMessagingHosts
