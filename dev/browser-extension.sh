#!/bin/bash

yarn

echo "Watching the browser extension and dependencies..."
yarn workspace browser-extensions run dev
